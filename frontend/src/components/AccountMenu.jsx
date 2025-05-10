import React, { useState } from "react";
import Button from "./Button";
import * as styles from "../App.module.css";

export default function AccountMenu({ username, email, onLogout }) {
  const [isOpen, setIsOpen] = useState(false);

  const toggleDropdown = () => setIsOpen(!isOpen);

  return (
    <div className={styles.accountMenu}>
      <Button
        text="My Account"
        onClick={toggleDropdown}
        className={styles.accountButton}
      />
      {isOpen && (
        <div className={styles.dropdown}>
          <div className={styles.dropdownItem}>{username}</div>
          <div className={styles.dropdownItem}>{email}</div>
          <button className={styles.logoutButton} onClick={onLogout}>
            Logout
          </button>
        </div>
      )}
    </div>
  );
}
