import { Button } from "@/components/common";
import { FormProps } from "./Form.types";
import { useForm } from "./useForm";
import cn from "classnames";
import { useEffect, useMemo, useRef } from "react";
import { AnimatedField } from "./components";

const Form = (props: FormProps) => {
  const formLogic = useForm({ ...props });

  // Memos
  const visibleFields = useMemo(
    () =>
      Object.values(formLogic.fieldsState).filter((field) => {
        if (field.display === undefined) return true;

        return typeof field.display === "function"
          ? field.display(formLogic.fieldsState)
          : field.display;
      }),
    [formLogic.fieldsState],
  );

  // Refs
  const mounted = useRef(false);

  // Effects
  useEffect(() => {
    mounted.current = true;
  }, []);

  return (
    <form className="form" noValidate>
      {Object.values(formLogic.fieldsState).map((field) => (
        <AnimatedField
          key={field.name}
          field={field}
          visible={visibleFields.some((visible) => visible.name === field.name)}
          formLogic={formLogic}
        />
      ))}

      <Button
        {...(props.submitButton || { value: "Submit" })}
        className={cn("form-button", props.submitButton?.className)}
        onClick={(e) => {
          e?.preventDefault();

          const isValid = formLogic.validate();

          if (isValid) {
            props.onSubmit(
              Object.fromEntries(
                Object.values(formLogic.fieldsState)
                  .filter(
                    (field) =>
                      !formLogic.isFieldDisabled(
                        field,
                        formLogic.fieldsState,
                      ) &&
                      formLogic.isFieldDisplayed(field, formLogic.fieldsState),
                  )
                  .map((field) => [field.name, field.value]),
              ),
            );
          }
        }}
        disabled={
          !!Object.values(formLogic.fieldsState).find(
            (field) =>
              !formLogic.isFieldDisabled(field, formLogic.fieldsState) &&
              field.error,
          )
        }
      />
    </form>
  );
};

export default Form;
