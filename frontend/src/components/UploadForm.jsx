import React, { useState } from "react";
import Button from "./Button";

export default function UploadForm({ onSuccess }) {
  const [file, setFile] = useState(null);
  const [isUploading, setIsUploading] = useState(false); // Track upload state

  const handleUpload = async () => {
    if (isUploading || !file) return; // Prevent multiple uploads
    setIsUploading(true);

    const formData = new FormData();
    formData.append("dataset", file);

    const res = await fetch("http://localhost:8080/datasets/upload", {
      method: "POST",
      credentials: "include",
      body: formData,
    });

    setIsUploading(false); // Re-enable button after upload

    if (res.ok) {
      onSuccess();
      setFile(null); // Reset file input after successful upload
    } else {
      const err = await res.json().catch(() => ({}));
      alert(err.error || "Upload failed");
    }
  };

  return (
    <>
      <input
        aria-label="Dataset Upload"
        type="file"
        onChange={(e) => setFile(e.target.files[0])}
      />
      <br />
      <Button
        text="Upload Dataset"
        onClick={handleUpload}
        disabled={isUploading || !file} // Disable button when uploading or no file selected
      />
    </>
  );
}
