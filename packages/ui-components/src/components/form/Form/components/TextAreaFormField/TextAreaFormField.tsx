import { ProgressBar } from "@/components/common";
import { TextAreaFormFieldProps } from "./TextAreaFormField.types";

const TextAreaFormField = (props: TextAreaFormFieldProps) => {
  return (
    <>
      <textarea
        className="form-field text-area-form-field"
        value={props.value}
        onChange={(e) => {
          let value = e.target.value;

          if (props.format) value = props.format(value);

          if (props.charsMax && value.length > props.charsMax) return;

          props.setValue(value);
        }}
        placeholder={props.placeholder}
        name={props.name}
        style={{ resize: props.resize ? "vertical" : "none" }}
      />
      {props.charsMax && (
        <div className="chars-check">
          <ProgressBar
            percentage={(props.value.length / props.charsMax) * 100}
          />
          <span
            style={{
              color:
                props.value.length >= props.charsMax
                  ? "var(--primary-color)"
                  : "var(--secondary-text-color)",
            }}
          >
            {props.value.length} / {props.charsMax} chars
          </span>
        </div>
      )}
    </>
  );
};

export default TextAreaFormField;
