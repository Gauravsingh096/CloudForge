import { Routes, Route, Link } from "react-router-dom";
import ProjectsPage from "./pages/ProjectsPage";
import DeploysPage from "./pages/DeploysPage";

export default function App() {
  return (
    <div style={{ fontFamily: "system-ui, sans-serif", maxWidth: 900, margin: "0 auto", padding: "1rem" }}>
      <header style={{ borderBottom: "1px solid #e5e7eb", paddingBottom: "1rem", marginBottom: "1.5rem" }}>
        <nav style={{ display: "flex", gap: "1.5rem", alignItems: "center" }}>
          <strong style={{ fontSize: "1.2rem" }}>⚡ CloudForge</strong>
          <Link to="/" style={linkStyle}>Projects</Link>
        </nav>
      </header>
      <main>
        <Routes>
          <Route path="/" element={<ProjectsPage />} />
          <Route path="/projects/:id/deploys" element={<DeploysPage />} />
        </Routes>
      </main>
    </div>
  );
}

const linkStyle: React.CSSProperties = {
  color: "#2563eb",
  textDecoration: "none",
  fontWeight: 500,
};
