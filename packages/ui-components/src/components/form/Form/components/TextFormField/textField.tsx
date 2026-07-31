import TextFormField from "./TextFormField";
import { FormField } from "../../Form.types";

interface TextFieldOptions {
  name: string;
  label: string;
  value?: string;
}

export const textField = ({
  name,
  label,
  value = "",
}: TextFieldOptions): FormField => {
  return {
    name,
    label,
    value,

    render({ value, setValue }) {
      return (
        <TextFormField
          value={value as string}
          setValue={setValue as (value: string) => void}
        />
      );
    },
  };
};
