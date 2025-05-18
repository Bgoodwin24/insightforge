import React, { useState } from 'react';
import Button from './Button';
import * as styles from './DropdownMenu.module.css';

export default function DropdownMenu({ label, options, disabled = false }) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className={styles.dropdown}>
      <Button
        text={label}
        onClick={() => !disabled && setIsOpen(!isOpen)}
        className={styles.dropdownButton}
        disabled={disabled}
      />
      {isOpen && !disabled && (
        <div className={styles.dropdownContent}>
          {options.map(({ text, onClick, children, disabled: childDisabled }, idx) =>
            children ? (
              <DropdownMenu
                key={idx}
                label={text}
                options={children}
                disabled={childDisabled}
              />
            ) : (
              <Button
                key={idx}
                text={text}
                onClick={childDisabled ? undefined : onClick}
                className={styles.itemButton}
                disabled={childDisabled}
              />
            )
          )}
        </div>
      )}
    </div>
  );
}
