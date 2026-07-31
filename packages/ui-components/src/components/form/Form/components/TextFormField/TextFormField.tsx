import { TextFormFieldProps } from "./TextFormField.types";

const TextFormField = (props: TextFormFieldProps) => {
  return (
    <input
      className="form-field text-form-field"
      type={props.masked ? "password" : "text"}
      value={props.value}
      name={props.name}
      onChange={(e) => props.setValue(e.target.value)}
      placeholder={props.placeholder}
    />
  );
};

export default TextFormField;
