import { CSSProperties, useEffect, useMemo, useState } from "react";
import { SelectFormFieldProps } from "./SelectFormField.types";
import cn from "classnames";

const SelectFormField = (props: SelectFormFieldProps) => {
  // States
  const [isOpen, setIsOpen] = useState<boolean>(false);

  // Effects
  useEffect(() => {
    if (props.value == "") props.setValue(null);
  }, [props]);

  // Memos
  const displayed = useMemo(() => {
    const choice = props.options.find((c) => c.value === props.value);
    return choice?.displayed ?? <>{choice?.value}</>;
  }, [props.options, props.value]);

  return (
    <div
      className={cn(
        "form-field select-form-field",
        isOpen && "open",
        props.disabled && "disabled",
      )}
    >
      <div
        className={cn("selected-box", !props.value && "placeholder")}
        onMouseDown={(e) => {
          e.preventDefault();
        }}
        onClick={(e) => {
          e.preventDefault();
          setIsOpen((prev) => !prev);
        }}
      >
        {props.value ? displayed : props.placeholder}
      </div>
      <span className="material-symbols-outlined arrow-icon">
        keyboard_arrow_down
      </span>
      <div
        className="choices"
        style={{ "--options": props.options.length } as CSSProperties}
      >
        {props.options.map((choice, i) => {
          return (
            <div
              className={cn(
                "choice",
                choice.value == props.value && "selected",
              )}
              onClick={() => {
                props.setValue(choice.value);
                setIsOpen(false);
              }}
              key={i}
            >
              {choice.displayed || choice.value}
              {choice.value == props.value && (
                <span className="material-symbols-outlined check-icon">
                  check
                </span>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default SelectFormField;
