import { JSX } from "react/jsx-runtime";
import { FormField, FormFieldType, FormProps } from "./Form.types";
import { useForm } from "./useForm";
import { TextFormField } from "./components";

const Form = (props: FormProps) => {
  const formLogic = useForm({ ...props });

  // Functions
  const renderField = (field: FormField): JSX.Element => {
    switch (field.type) {
      case FormFieldType.Text:
        return <TextFormField field={field} />;
      default:
        return <></>;
    }
  };

  return (
    <form className="form">
      {formLogic.fields.map((field, i) => {
        return (
          <div className="form-field-wrapper" key={i}>
            <label htmlFor={field.name}>{field.label}</label>
            {renderField(field)}
          </div>
        );
      })}
    </form>
  );
};

export default Form;
