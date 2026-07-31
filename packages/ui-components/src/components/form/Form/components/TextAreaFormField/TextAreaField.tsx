import TextAreaFormField from "./TextAreaFormField";
import { FormField } from "../../Form.types";
import { TextAreaFormFieldProps } from "./TextAreaFormField.types";

interface TextAreaFieldOptions extends Omit<
  TextAreaFormFieldProps,
  "render" | "setValue"
> {}

export const textAreaField = (props: TextAreaFieldOptions): FormField => {
  return {
    ...props,

    render: (props: TextAreaFormFieldProps) => {
      return <TextAreaFormField {...props} />;
    },
  };
};
