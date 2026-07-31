import TextFormField from "./TextFormField";
import { FormField } from "../../Form.types";
import { TextFormFieldProps } from "./TextFormField.types";

interface TextFieldOptions extends Omit<
  TextFormFieldProps,
  "render" | "setValue"
> {}

export const textField = (props: TextFieldOptions): FormField => {
  return {
    ...props,

    render: (props: TextFormFieldProps) => {
      return <TextFormField {...props} />;
    },
  };
};
