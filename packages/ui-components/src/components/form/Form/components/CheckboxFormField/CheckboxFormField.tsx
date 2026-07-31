import { useEffect } from "react";
import { CheckboxFormFieldProps } from "./CheckboxFormField.types";

const CheckboxFormField = (props: CheckboxFormFieldProps) => {
  // Effects
  useEffect(() => {
    if (Array.isArray(props.value) && props.value.length === 0)
      props.setValue(null);
  }, [props]);

  return (
    <div className="form-field checkbox-form-field grid">
      {props.options.map((opt, i) => {
        return (
          <div
            className="choice"
            key={i}
            onClick={() => {
              if (!props.value) props.setValue([opt.value]);
              else if (!props.value.includes(opt.value)) {
                props.setValue([...props.value, opt.value]);
              } else {
                const tmp = [...props.value];
                tmp.splice(tmp.indexOf(opt.value), 1);

                props.setValue(tmp);
              }
            }}
          >
            <input
              className="checkbox-option"
              type="checkbox"
              checked={
                Array.isArray(props.value) && props.value.includes(opt.value)
              }
              placeholder={props.placeholder}
              name={props.name}
            />
            <span className="material-symbols-outlined checkbox-icon">
              {Array.isArray(props.value) && props.value.includes(opt.value)
                ? "check_box"
                : "check_box_outline_blank"}
            </span>
            <span>{opt.displayed || opt.value}</span>
          </div>
        );
      })}
    </div>
  );
};

export default CheckboxFormField;
