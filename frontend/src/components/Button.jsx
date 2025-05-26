import React from "react";
import * as styles from "./Button.module.css";

export default function Button({ text, onClick, className, disabled = false }) {
    if (!styles || !styles.button) {
    console.error("Styles object or button class is missing");
  }

    return (
        <button className={`${styles.button} ${className || ""}`} onClick={onClick} disabled={disabled}>
            {text}
        </button>
    );
}
