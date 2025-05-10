import React, { useState } from "react";
import Button from "./components/Button";
import Modal from "./components/Modal";
import AuthForm from "./components/AuthForm";
import UploadForm from "./components/UploadForm";
import AccountMenu from "./components/AccountMenu";
import * as styles from "./App.module.css";

export default function App() {
  const [modalType, setModalType] = useState(null);
  const [user, setUser] = useState(null); // Track the logged-in user

  const closeModal = () => setModalType(null);

  const handleLogout = async () => {
    await fetch("http://localhost:8080/user/logout", {
      method: "POST",
      credentials: "include",
    });
    setUser(null);
  };

  return (
    <div className={styles.app}>
      {/* Header */}
      <header className={styles.header}>
        <div className={styles.logoContainer}>
          <img
            src="/assets/InsightForge_Logo.png"
            alt="InsightForge Logo"
            className={styles.logo}
          />
        </div>

        <div className={styles.titles}>
          <h1 className={styles.title}>InsightForge</h1>
          <h2 className={styles.title2}>Visualize your data, forge your own path</h2>
        </div>

        <div className={styles.buttonContainer}>
          {!user ? (
            <Button text="Login / Register" onClick={() => setModalType("auth")} />
          ) : (
            <div className={styles.accountMenu}>
              <AccountMenu
                username={user.username}
                email={user.email}
                onLogout={handleLogout}
              />
            </div>
          )}
          <Button text="Upload Dataset" onClick={() => setModalType("upload")} />
        </div>
      </header>

      <div className={styles.divider}></div>

      {/* Main Content */}
      <div className={styles.mainContent}>
        <p>Explore your datasets here.</p>

        {modalType === "auth" && (
          <Modal title="Login / Register" onClose={closeModal}>
            <AuthForm
              onSuccess={(userData) => {
                setUser(userData); // Set user after login/register
                closeModal();
              }}
            />
          </Modal>
        )}

        {modalType === "upload" && (
          <Modal title="Upload Dataset" onClose={closeModal}>
            <UploadForm onSuccess={closeModal} />
          </Modal>
        )}
      </div>
    </div>
  );
}
