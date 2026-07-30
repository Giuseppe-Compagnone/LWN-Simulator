import { useState } from "react";
import { FormField, FormLogic, UseFormProps } from "./Form.types";

export const useForm = (props: UseFormProps): FormLogic => {
  // States
  const [fields, setFields] = useState<Array<FormField>>(props.fields);

  return {
    fields,
    setFields,
  };
};
