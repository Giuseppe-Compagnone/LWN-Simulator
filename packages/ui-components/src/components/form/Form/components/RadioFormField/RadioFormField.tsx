import { useEffect } from "react";
import { RadioFormFieldProps } from "./RadioFormField.types";

const RadioFormField = (props: RadioFormFieldProps) => {
  // Effects
  useEffect(() => {
    if (props.value == "") props.setValue(null);
  }, [props]);

  return (
    <div className="form-field radio-form-field grid">
      {props.options.map((opt, i) => {
        return (
          <div
            className="choice"
            key={i}
            onClick={() => {
              props.setValue(opt.value);
            }}
          >
            <input
              className="radio-option"
              type="radio"
              checked={props.value == opt.value}
              readOnly
              placeholder={props.placeholder}
              name={props.name}
            />
            <span className="material-symbols-outlined radio-icon">
              {props.value == opt.value
                ? "radio_button_checked"
                : "radio_button_unchecked"}
            </span>
            <span>{opt.displayed || opt.value}</span>
          </div>
        );
      })}
    </div>
  );
};

export default RadioFormField;
