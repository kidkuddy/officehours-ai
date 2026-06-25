import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import App from "./App";
import { applyTheme, resolveInitialTheme } from "./hooks/useTheme";
import "./styles/tokens.css";
import "./styles/layout.css";
import "./styles/app.css";

// Apply the persisted/OS theme before first paint to avoid a flash.
applyTheme(resolveInitialTheme());

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </React.StrictMode>,
);
