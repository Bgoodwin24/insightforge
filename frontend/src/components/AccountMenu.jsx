import React, { useState, useEffect } from "react";
import Button from "./Button";
import * as styles from "../App.module.css";

export default function AccountMenu({ onLogout }) {
  const [isOpen, setIsOpen] = useState(false);
  const [user, setUser] = useState(null);    // State to store fetched profile data
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const toggleDropdown = () => setIsOpen(!isOpen);

  useEffect(() => {
    // Fetch user profile on mount
    async function fetchProfile() {
      try {
        const res = await fetch("http://localhost:8080/user/profile", {
          credentials: "include",  // Send cookies/session tokens
        });

        if (!res.ok) {
          throw new Error("Failed to fetch profile");
        }

        const data = await res.json();
        setUser(data.user);
        setError(null);
      } catch (err) {
        setUser(null);
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }

    fetchProfile();
  }, []);

  if (loading) return <div>Loading account info...</div>;
  if (error) return <div>Error loading account info</div>;
  if (!user) return null;  // Or render something if user is not logged in

  return (
    <div className={styles.accountMenu}>
      <Button
        text="My Account"
        onClick={toggleDropdown}
        className={styles.accountButton}
      />
      {isOpen && (
        <div className={styles.dropdown}>
          <div className={styles.dropdownItem}>{user.username}</div>
          <div className={styles.dropdownItem}>{user.email}</div>
          <button disabled={!user} className={styles.logoutButton} onClick={onLogout}>
            Logout
          </button>
        </div>
      )}
    </div>
  );
}
  