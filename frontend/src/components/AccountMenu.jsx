import React, { useState } from "react";
import Button from "./Button";
import * as styles from "../App.module.css";

export default function AccountMenu({ user, onLogout }) {
  const [isOpen, setIsOpen] = useState(false);
  const toggleDropdown = () => setIsOpen(!isOpen);

  if (!user) return null; // Don't render if no user is logged in

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
