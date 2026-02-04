from airflow.operators.python import PythonOperator
from clickhouse_driver import Client
from datetime import datetime

from airflow import DAG


# ETL функция
def etl_user_reports():
    try:
        sensor_data = [
            {'user_id': 'user1', 'value': 23, 'ts': '2026-02-04 00:47:00'},
            {'user_id': 'user2', 'value': 17, 'ts': '2026-02-04 01:12:00'},
            {'user_id': 'admin1', 'value': 42, 'ts': '2026-02-04 00:05:00'},
            {'user_id': 'prothetic1', 'value': 9, 'ts': '2026-02-04 00:33:00'},
            {'user_id': 'prothetic2', 'value': 35, 'ts': '2026-02-04 00:21:00'},
            {'user_id': 'prothetic3', 'value': 28, 'ts': '2026-02-04 00:58:00'}
        ]

        crm_data = [
            {'user_id': 'user1', 'plan': 'basic'},
            {'user_id': 'user2', 'plan': 'basic'},
            {'user_id': 'admin1', 'plan': 'enterprise'},
            {'user_id': 'prothetic1', 'plan': 'prothetic_plan'},
            {'user_id': 'prothetic2', 'plan': 'prothetic_plan'},
            {'user_id': 'prothetic3', 'plan': 'prothetic_plan'}
        ]

        merged_data = []
        for s in sensor_data:
            crm = next((c for c in crm_data if c["user_id"] == s["user_id"]), {})
            merged_data.append((
                s["user_id"],
                s["value"],
                datetime.strptime(s["ts"], "%Y-%m-%d %H:%M:%S"),
                crm.get("plan", None)
            ))

        client = Client(host='clickhouse', user='default', password='', port=9000)
        client.execute(
            'INSERT INTO user_reports (user_id, sensor_value, ts, plan) VALUES',
            merged_data
        )
    except Exception as e:
        # Логируем ошибку и пробрасываем её дальше
        print(f"Ошибка ETL: {e}")
        raise

# DAG
with DAG(
        dag_id="etl_user_reports",
        start_date=datetime(2026, 2, 4),
        schedule_interval="*/5 * * * *",  # каждые 5 минут
        catchup=False,
        tags=["reports", "etl"],
) as dag:

    run_etl = PythonOperator(
        task_id="run_etl",
        python_callable=etl_user_reports
    )