import { useState } from "react";
import { FormField, FormLogic, FormValue } from "./Form.types";

interface UseFormProps {
  fields: FormField[];
}

export const useForm = (props: UseFormProps): FormLogic => {
  const [values, setValues] = useState<Record<string, FormValue>>(() =>
    Object.fromEntries(props.fields.map((field) => [field.name, field.value])),
  );

  const setValue = (name: string, value: FormValue) => {
    setValues((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  return {
    values,
    setValue,
  };
};
