import React, { useState } from "react";
import Button from "./Button";

export default function AuthForm({ mode = "login", onSuccess }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [username, setUsername] = useState(""); // State for the username
  const [formMode, setFormMode] = useState(mode);
  const [isSubmitting, setIsSubmitting] = useState(false); // Track submission state

  const toggleMode = () => setFormMode(formMode === "login" ? "register" : "login");

  const handleSubmit = async () => {
    if (isSubmitting) return; // Prevent multiple submissions
    setIsSubmitting(true);

    const endpoint = formMode === "login" ? "/user/login" : "/user/register";

    const body = formMode === "login"
      ? { email, password }
      : { email, password, username }; // Include username when registering

    const res = await fetch(`http://localhost:8080${endpoint}`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });

    setIsSubmitting(false); // Re-enable button after submission

  if (res.ok) {
    const data = await res.json().catch(() => ({}));

    // Use returned data if present, otherwise fallback to form inputs
    onSuccess({
      username: data.username || username,
      email: data.email || email,
    });
  } else {
      const err = await res.json().catch(() => ({}));
      alert(err.error || `${formMode} failed`);
    }
  };

  const isFormValid =
  email.trim() !== "" &&
  password.trim() !== "" &&
  (formMode === "login" || username.trim() !== "");

  return (
    <>
      {formMode === "register" && (
        <input
          aria-label="Username"
          placeholder="Username"
          type="text"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
      )}
      <br />
      <input
        aria-label="Email"
        placeholder="Email"
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
      />
      <br />
      <input
        aria-label="Password"
        placeholder="Password"
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
      />
      <br />
      <Button text={formMode === "login" ? "Login" : "Register"} onClick={handleSubmit} disabled={isSubmitting || !isFormValid} />
      <p onClick={toggleMode} style={{ cursor: "pointer", marginTop: "10px" }}>
        {formMode === "login"
          ? "Need an account? Register"
          : "Have an account? Login"}
      </p>
    </>
  );
}
