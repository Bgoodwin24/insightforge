import React from "react";
import * as styles from "./Button.module.css";

console.log("Button Styles", styles);

export default function Button({ text, onClick, className}) {
    if (!styles || !styles.button) {
    console.error("Styles object or button class is missing");
  }

    return (
        <button className={`${styles.button} ${className || ""}`} onClick={onClick}>
            {text}
        </button>
    );
}
