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
      const field = prev[name];
      field.value = value;
      field.error = null;

      if (field.validations && field.value) {
        for (let i = 0; i < field.validations.length; i++) {
          const validation = field.validations[i];

          if (!validation.realtime) continue;

          const isValid = validation.rule.test(field.value as string);
          if (!isValid) {
            field.error = validation.error;
            break;
          }
        }
      }

      return { ...prev };
    });
  };

  const validate = () => {
    const tmp = { ...fieldsState };
    let isValid = true;

    Object.values(tmp).forEach((field) => {
      if (!field.disabled) {
        if (field.required && !field.value) {
          isValid = false;
          field.error = "This field is required";
        }

        if (field.validations && isValid) {
          for (let i = 0; i < field.validations.length; i++) {
            const validation = field.validations[i];

            isValid = validation.rule.test((field.value as string) || "");
            if (!isValid) {
              field.error = validation.error;
              break;
            }
          }
        }
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
