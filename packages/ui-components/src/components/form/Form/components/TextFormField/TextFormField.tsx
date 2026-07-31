import { TextFormFieldProps } from "./TextFormField.types";

const TextFormField = (props: TextFormFieldProps) => {
  return (
    <input
      value={props.value}
      onChange={(e) => props.setValue(e.target.value)}
    />
  );
};

export default TextFormField;
