import { ProgressBar } from "@/components/common";
import { TextAreaFormFieldProps } from "./TextAreaFormField.types";

const TextAreaFormField = (props: TextAreaFormFieldProps) => {
  return (
    <>
      <textarea
        className="form-field text-area-form-field"
        value={props.value}
        onChange={(e) => {
          if (props.charsMax && e.target.value.length > props.charsMax) return;

          props.setValue(e.target.value);
        }}
        placeholder={props.placeholder}
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
