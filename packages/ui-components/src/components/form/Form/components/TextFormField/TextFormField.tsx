import { TextFormFieldProps } from "./TextFormField.types";

const TextFormField = (props: TextFormFieldProps) => {
  return (
    <input
      className="text-form-field"
      type="text"
      value={props.value}
      onChange={(e) => props.setValue(e.target.value)}
    />
  );
};

export default TextFormField;
