import cn from "classnames";
import { CardLayout, CardProps, CardType } from "./Card.types";

const Card = (props: CardProps) => {
  const layout = props.layout || CardLayout.Default;
  const type = props.type || CardType.Default;

  return (
    <div className={cn("card", props.className, layout, type)}>
      {props.children}
    </div>
  );
};

export default Card;
