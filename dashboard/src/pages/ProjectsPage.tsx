import useSWR from "swr";
import { Link } from "react-router-dom";

interface Project {
  id: string;
  name: string;
  repo_url: string;
  subdomain: string;
  created_at: string;
}

const fetcher = (url: string) => fetch(url).then((r) => r.json());

export default function ProjectsPage() {
  const { data: projects, error, isLoading } = useSWR<Project[]>("/api/projects", fetcher);

  if (isLoading) return <p>Loading projects…</p>;
  if (error) return <p style={{ color: "red" }}>Failed to load projects.</p>;

  return (
    <div>
      <h1 style={{ fontSize: "1.4rem", marginBottom: "1rem" }}>Projects</h1>
      {projects?.length === 0 && <p>No projects yet.</p>}
      <ul style={{ listStyle: "none", padding: 0 }}>
        {projects?.map((p) => (
          <li key={p.id} style={cardStyle}>
            <div>
              <strong>{p.name}</strong>
              <span style={{ marginLeft: "0.75rem", color: "#6b7280", fontSize: "0.875rem" }}>
                {p.subdomain}.cloudforge.is-a.dev
              </span>
            </div>
            <div style={{ marginTop: "0.25rem", fontSize: "0.875rem", color: "#6b7280" }}>
              {p.repo_url}
            </div>
            <div style={{ marginTop: "0.5rem" }}>
              <Link to={`/projects/${p.id}/deploys`} style={{ fontSize: "0.875rem", color: "#2563eb" }}>
                View deploys →
              </Link>
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}

const cardStyle: React.CSSProperties = {
  border: "1px solid #e5e7eb",
  borderRadius: "0.5rem",
  padding: "1rem",
  marginBottom: "0.75rem",
};
