import React, { useState } from "react";
import Button from "./Button";

export default function UploadForm({ onSuccess, user }) {
  const [file, setFile] = useState(null);
  const [isUploading, setIsUploading] = useState(false);

  const handleUpload = async () => {
    if (isUploading || !file || !user) return;
    setIsUploading(true);

    const formData = new FormData();
    formData.append("dataset", file);

    const res = await fetch("http://localhost:8080/datasets/upload", {
      method: "POST",
      credentials: "include",
      body: formData,
    });

    setIsUploading(false);

    if (res.ok) {
      onSuccess();
      setFile(null);
    } else {
      const err = await res.json().catch(() => ({}));
      alert(err.error || "Upload failed");
    }
  };

  const isDisabled = isUploading || !file || !user;
  const tooltip = !user ? "Login to upload a dataset" : "";

  return (
    <>
      {/* Custom-styled file input */}
      <label>
        <input
          type="file"
          accept=".csv"
          onChange={(e) => setFile(e.target.files[0])}
          style={{ display: "none" }}
        />
        <Button text={file ? `File: ${file.name}` : "Choose File"} />
      </label>

      <br />

      <Button
        text="Upload Dataset"
        onClick={handleUpload}
        disabled={isDisabled}
        title={isDisabled ? tooltip : ""}
      />
    </>
  );
}
