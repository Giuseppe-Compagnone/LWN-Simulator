import { Button } from "@/components/common";
import { FormProps } from "./Form.types";
import { useForm } from "./useForm";

const Form = (props: FormProps) => {
  const formLogic = useForm({ ...props });

  return (
    <form className="form">
      {props.fields.map((field) => (
        <div className="form-field-wrapper" key={field.name}>
          <label htmlFor={field.name}>{field.label}</label>

          {field.render({
            value: formLogic.values[field.name],
            setValue: (v) => formLogic.setValue(field.name, v),
          })}
        </div>
      ))}
      <Button
        {...(props.button || { value: "Submit" })}
        onClick={(e) => {
          e?.preventDefault();
          props.onSubmit(formLogic.values);
        }}
      />
    </form>
  );
};

export default Form;
