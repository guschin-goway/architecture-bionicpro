import React, { useState } from 'react';
import { useKeycloak } from '@react-keycloak/web';

interface SensorReport {
  user_id: string;
  sensor_value: string; // или number, если будете конвертировать
  ts: string;
  plan: string;
}

const ReportPage: React.FC = () => {
  const { keycloak, initialized } = useKeycloak();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [reports, setReports] = useState<SensorReport[]>([]);

  const downloadReport = async () => {
    if (!keycloak?.token) {
      setError('Not authenticated');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const response = await fetch(`${process.env.REACT_APP_API_URL}/reports`, {
        headers: {
          'Authorization': `Bearer ${keycloak.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error ${response.status}`);
      }

      const data = await response.json();
      if (!Array.isArray(data)) {
        throw new Error("Данные отсутствуют");
      }

      setReports(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
      setReports([]);
    } finally {
      setLoading(false);
    }
  };

  if (!initialized) return <div>Loading...</div>;

  if (!keycloak.authenticated) {
    return (
        <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100">
          <button
              onClick={() => keycloak.login()}
              className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            Login
          </button>
        </div>
    );
  }

  return (
      <div className="flex flex-col items-center justify-start min-h-screen bg-gray-100 p-8">
        <div className="w-full max-w-4xl p-8 bg-white rounded-lg shadow-md">
          <h1 className="text-2xl font-bold mb-6">Usage Reports</h1>

          <button
              onClick={downloadReport}
              disabled={loading}
              className={`px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 ${
                  loading ? 'opacity-50 cursor-not-allowed' : ''
              }`}
          >
            {loading ? 'Generating Report...' : 'Download Report'}
          </button>

          {error && (
              <div className="mt-4 p-4 bg-red-100 text-red-700 rounded">
                {error}
              </div>
          )}

          {reports.length > 0 && (
              <table className="mt-6 w-full border-collapse border border-gray-300">
                <thead>
                <tr className="bg-gray-100">
                  <th className="border border-gray-300 px-4 py-2">User ID</th>
                  <th className="border border-gray-300 px-4 py-2">Sensor Value</th>
                  <th className="border border-gray-300 px-4 py-2">Timestamp</th>
                  <th className="border border-gray-300 px-4 py-2">Plan</th>
                </tr>
                </thead>
                <tbody>
                {reports.map((r, idx) => (
                    <tr key={idx} className="text-center">
                      <td className="border border-gray-300 px-4 py-2">{r.user_id}</td>
                      <td className="border border-gray-300 px-4 py-2">{r.sensor_value}</td>
                      <td className="border border-gray-300 px-4 py-2">{r.ts}</td>
                      <td className="border border-gray-300 px-4 py-2">{r.plan}</td>
                    </tr>
                ))}
                </tbody>
              </table>
          )}
        </div>
      </div>
  );
};

export default ReportPage;