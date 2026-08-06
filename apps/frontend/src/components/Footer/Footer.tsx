import { FooterProps } from "./Footer.types";

const Footer = (props: FooterProps) => {
  return (
    <footer className="footer">
      <div className="version">BUILD: v2.0.0</div>
      <div className="repo">
        <a
          href="https://github.com/Giuseppe-Compagnone/LWN-Simulator"
          target="_blank"
        >
          Repository
        </a>
      </div>
    </footer>
  );
};

export default Footer;
