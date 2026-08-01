import { useState } from "react";
import { FormField, FormLogic, FormValue } from "./Form.types";

interface UseFormProps {
  fields: FormField[];
}

export const useForm = (props: UseFormProps): FormLogic => {
  // States
  const [fieldsState, setFieldsState] = useState<Record<string, FormField>>(
    () => Object.fromEntries(props.fields.map((field) => [field.name, field])),
  );

  // Functions
  const setValue = (name: string, value: FormValue) => {
    setFieldsState((prev) => {
      prev[name].value = value;
      prev[name].error = null;

      return { ...prev };
    });
  };

  const validate = () => {
    const tmp = { ...fieldsState };
    let isValid = true;

    Object.values(tmp).forEach((field) => {
      if (field.required && !field.value) {
        isValid = false;
        field.error = "This field is required";
        console.log(field.name, field.error);
      }
    });

    if (!isValid) setFieldsState(tmp);

    return isValid;
  };

  return {
    fieldsState,
    setValue,
    validate,
  };
};
