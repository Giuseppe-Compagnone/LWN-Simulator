"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { FormField, FormLogic, FormValue, UseFormProps } from "./Form.types";

export const useForm = (props: UseFormProps): FormLogic => {
  // States
  const [fieldsState, setFieldsState] = useState<Record<string, FormField>>(
    () => Object.fromEntries(props.fields.map((field) => [field.name, field])),
  );

  // Refs
  const fieldsStateRef = useRef<Record<string, FormField>>(
    Object.fromEntries(props.fields.map((field) => [field.name, field])),
  );
  const formLogicRef = useRef<FormLogic | null>(null);

  // Functions
  const setValue = useCallback((name: string, value: FormValue) => {
    const prev = fieldsStateRef.current;
    const field = prev[name];
    const nextValue = field.format ? field.format(value as string) : value;

    if (field.value === nextValue && field.error === null) return;

    const updatedField: FormField = {
      ...field,
      value: nextValue,
      error: null,
    };

    if (updatedField.validations && updatedField.value) {
      for (const validation of updatedField.validations) {
        if (!validation.realtime) continue;

        const isValid = validation.rule.test(updatedField.value as string);

        if (!isValid) {
          updatedField.error = validation.error;
          break;
        }
      }
    }

    const nextState = {
      ...prev,
      [name]: updatedField,
    };

    fieldsStateRef.current = nextState;
    setFieldsState(nextState);

    if (updatedField.onChange) {
      if (formLogicRef.current) {
        updatedField.onChange({
          ...formLogicRef.current,
          fieldsState: nextState,
        });
      }
    }
  }, []);

  const setProp = useCallback((name: string, prop: string, value: unknown) => {
    const prev = fieldsStateRef.current;
    const field = prev[name];

    if (!field) return;

    const updatedField: FormField = {
      ...field,
      [prop]: value,
    };

    const nextState = {
      ...prev,
      [name]: updatedField,
    };

    fieldsStateRef.current = nextState;
    setFieldsState(nextState);
  }, []);

  const isFieldDisabled = useCallback((
    field: FormField,
    fieldsState: Record<string, FormField>,
  ) => {
    return typeof field.disabled === "function"
      ? field.disabled(fieldsState)
      : !!field.disabled;
  }, []);

  const isFieldDisplayed = useCallback((
    field: FormField,
    fieldsState: Record<string, FormField>,
  ) => {
    if (typeof field.display === "function") {
      return field.display(fieldsState);
    }

    return field.display ?? true;
  }, []);

  useEffect(() => {
    setFieldsState((currentState) => {
      const hasConfigurationChanged =
        Object.keys(currentState).length !== props.fields.length ||
        props.fields.some((field) => {
          const currentField = currentState[field.name];

          return !currentField ||
            currentField.label !== field.label ||
            currentField.placeholder !== field.placeholder ||
            currentField.required !== field.required ||
            currentField.disabled !== field.disabled ||
            currentField.readOnly !== field.readOnly ||
            currentField.display !== field.display ||
            currentField.info !== field.info ||
            currentField.toolbar !== field.toolbar;
        });

      if (!hasConfigurationChanged) return currentState;

      const nextState = Object.fromEntries(
        props.fields.map((field) => {
          const currentField = currentState[field.name];

          return [
            field.name,
            {
              ...field,
              value: currentField?.value ?? field.value,
              error: currentField?.error ?? field.error,
            },
          ];
        }),
      );

      fieldsStateRef.current = nextState;
      return nextState;
    });
  }, [props.fields]);

  const validate = useCallback(() => {
    const tmp: Record<string, FormField> = Object.fromEntries(
      Object.entries(fieldsState).map(([name, field]) => [
        name,
        { ...field, error: null },
      ]),
    );
    let isValid = true;

    Object.values(tmp).forEach((field) => {
      if (
        !isFieldDisabled(field, fieldsState) &&
        isFieldDisplayed(field, fieldsState)
      ) {
        let fieldIsValid = true;

        if (field.required && !field.value) {
          fieldIsValid = false;
          field.error = "This field is required";
        }

        if (field.validations && field.value && fieldIsValid) {
          for (let i = 0; i < field.validations.length; i++) {
            const validation = field.validations[i];

            fieldIsValid = validation.rule.test((field.value as string) || "");
            if (!fieldIsValid) {
              field.error = validation.error;
              break;
            }
          }
        }

        if (!fieldIsValid) isValid = false;
      }
    });

    if (!isValid) {
      fieldsStateRef.current = tmp;
      setFieldsState(tmp);
    }

    return isValid;
  }, [fieldsState, isFieldDisabled, isFieldDisplayed]);

  const formLogic = useMemo<FormLogic>(() => {
    return {
      fieldsState,
      setValue,
      setProp,
      isFieldDisabled,
      isFieldDisplayed,
      validate,
    };
  }, [
    fieldsState,
    setValue,
    setProp,
    isFieldDisabled,
    isFieldDisplayed,
    validate,
  ]);

  useEffect(() => {
    formLogicRef.current = formLogic;
  }, [formLogic]);

  return formLogic;
};
