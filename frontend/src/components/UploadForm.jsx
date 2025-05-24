import React, { useState, useRef } from "react";
import Button from "./Button";
import Modal from "./Modal";
import * as styles from "./UploadForm.module.css";

export default function UploadForm({ onSuccess, onDelete, onClose, user, datasetUploaded }) {
  const [file, setFile] = useState(null);
  const [isUploading, setIsUploading] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);

  const fileInputRef = useRef(null);

  const handleUpload = async () => {
    if (isUploading || !file || !user) return;
    setIsUploading(true);

    const formData = new FormData();
    formData.append("file", file);

    const res = await fetch("http://localhost:8080/datasets/upload", {
      method: "POST",
      credentials: "include",
      body: formData,
    });

    setIsUploading(false);

    if (res.ok) {
      const datasetInfo = await res.json();
      onSuccess(datasetInfo);
      setFile(null);
    } else {
      const err = await res.json().catch(() => ({}));
      alert(err.error || "Upload failed");
    }
  };

  const handleDelete = async () => {
    setShowDeleteModal(false);
    onDelete();
  };

  const isUploadDisabled = isUploading || !file || !user;
  const tooltip = !user ? "Login to upload a dataset" : "";

  return (
    <>
      {datasetUploaded ? (
        <>
          <Button text="Delete Dataset" onClick={() => setShowDeleteModal(true)} />
          {showDeleteModal && (
            <Modal onClose={() => setShowDeleteModal(false)}>
              <p>Are you sure you want to delete your dataset?</p>
              <div style={{ display: "flex", gap: "1rem", marginTop: "1rem" }}>
                <Button text="Yes" onClick={handleDelete} />
                <Button text="No" onClick={() => setShowDeleteModal(false)} />
              </div>
            </Modal>
          )}
        </>
      ) : isUploading ? (
        <div className={styles.uploading}>
          <p>
            Uploading dataset... <span className={styles.spinner}>⏳</span>
          </p>
        </div>
      ) : (
        <>
          <input
            ref={fileInputRef}
            type="file"
            accept=".csv"
            onChange={(e) => setFile(e.target.files[0])}
            style={{ display: "none" }}
          />

          <Button
            text={file ? `File: ${file.name}` : "Choose File"}
            onClick={() => fileInputRef.current?.click()}
          />

          <br />

          <Button
            text="Upload Dataset"
            onClick={handleUpload}
            disabled={isUploadDisabled}
            title={isUploadDisabled ? tooltip : ""}
          />
        </>
      )}
    </>
  );
}
