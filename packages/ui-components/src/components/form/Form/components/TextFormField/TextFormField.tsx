import { TextFormFieldProps } from "./TextFormField.types";

const TextFormField = (props: TextFormFieldProps) => {
  return (
    <input type="text" name={props.field.name} value={props.field.value} />
  );
};

export default TextFormField;
