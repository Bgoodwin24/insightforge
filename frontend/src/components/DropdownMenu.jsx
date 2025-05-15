import React, { useState } from 'react';
import Button from './Button';
import * as styles from './DropdownMenu.module.css';

export default function DropdownMenu({ label, options }) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className={styles.dropdown}>
      <Button
        text={label}
        onClick={() => setIsOpen(!isOpen)}
        className={styles.dropdownButton}
      />
      {isOpen && (
        <div className={styles.dropdownContent}>
          {options.map(({ text, onClick, children }, idx) =>
            children ? (
              <DropdownMenu key={idx} label={text} options={children} />
            ) : (
              <Button
                key={idx}
                text={text}
                onClick={onClick}
                className={styles.itemButton}
              />
            )
          )}
        </div>
      )}
    </div>
  );
}
