import useSWR from "swr";
import { useParams } from "react-router-dom";

interface Deploy {
  id: string;
  project_id: string;
  image: string;
  commit_sha: string;
  status: string;
  created_at: string;
}

const STATUS_COLOR: Record<string, string> = {
  pending: "#d97706",
  building: "#2563eb",
  deploying: "#7c3aed",
  running: "#16a34a",
  failed: "#dc2626",
  rolled_back: "#6b7280",
};

const fetcher = (url: string) => fetch(url).then((r) => r.json());

export default function DeploysPage() {
  const { id } = useParams<{ id: string }>();
  const { data: deploys, error, isLoading } = useSWR<Deploy[]>(
    `/api/projects/${id}/deploys`,
    fetcher,
    { refreshInterval: 5000 }
  );

  if (isLoading) return <p>Loading deploys…</p>;
  if (error) return <p style={{ color: "red" }}>Failed to load deploys.</p>;

  return (
    <div>
      <h1 style={{ fontSize: "1.4rem", marginBottom: "1rem" }}>Deploys</h1>
      {deploys?.length === 0 && <p>No deploys yet.</p>}
      <table style={{ width: "100%", borderCollapse: "collapse", fontSize: "0.9rem" }}>
        <thead>
          <tr style={{ textAlign: "left", borderBottom: "2px solid #e5e7eb" }}>
            <th style={thStyle}>Status</th>
            <th style={thStyle}>Commit</th>
            <th style={thStyle}>Image</th>
            <th style={thStyle}>Deployed at</th>
          </tr>
        </thead>
        <tbody>
          {deploys?.map((d) => (
            <tr key={d.id} style={{ borderBottom: "1px solid #f3f4f6" }}>
              <td style={tdStyle}>
                <span style={{
                  color: STATUS_COLOR[d.status] ?? "#374151",
                  fontWeight: 600,
                  textTransform: "capitalize",
                }}>
                  {d.status}
                </span>
              </td>
              <td style={tdStyle}>
                <code>{d.commit_sha.slice(0, 7) || "—"}</code>
              </td>
              <td style={{ ...tdStyle, color: "#6b7280", fontSize: "0.8rem" }}>
                {d.image.split(":")[1] ?? d.image}
              </td>
              <td style={tdStyle}>
                {new Date(d.created_at).toLocaleString()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

const thStyle: React.CSSProperties = { padding: "0.5rem 0.75rem" };
const tdStyle: React.CSSProperties = { padding: "0.6rem 0.75rem" };
