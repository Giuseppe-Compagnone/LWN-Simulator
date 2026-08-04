import cn from "classnames";
import { TextFormFieldProps } from "./TextFormField.types";

const TextFormField = (props: TextFormFieldProps) => {
  return (
    <input
      className={cn("form-field text-form-field", props.disabled && "disabled")}
      type={props.masked ? "password" : "text"}
      value={props.value}
      name={props.name}
      spellCheck={false}
      disabled={props.disabled}
      onChange={(e) => {
        let value = e.target.value;

        if (props.format) value = props.format(value);

        props.setValue(value);
      }}
      placeholder={props.placeholder}
    />
  );
};

export default TextFormField;
