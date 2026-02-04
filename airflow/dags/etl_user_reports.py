from airflow import DAG
from airflow.providers.http.operators.http import SimpleHttpOperator
from datetime import datetime

with DAG(
        dag_id="etl_user_reports",
        start_date=datetime(2024, 1, 1),
        schedule_interval="0 2 * * *",
        catchup=False,
        tags=["reports", "etl"],
) as dag:

    run_etl = SimpleHttpOperator(
        task_id="run_go_etl",
        http_conn_id="etl_service",
        endpoint="/etl/run",
        method="POST",
    )

run_etl

