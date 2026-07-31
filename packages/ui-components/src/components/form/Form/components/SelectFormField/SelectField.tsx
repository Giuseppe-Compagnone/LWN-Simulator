import SelectFormField from "./SelectFormField";
import { FormField } from "../../Form.types";
import { SelectFormFieldProps } from "./SelectFormField.types";

interface SelectFieldOptions extends Omit<
  SelectFormFieldProps,
  "render" | "setValue"
> {}

export const selectField = (props: SelectFieldOptions): FormField => {
  return {
    ...props,

    render: (props: SelectFormFieldProps) => {
      return <SelectFormField {...props} />;
    },
  };
};
