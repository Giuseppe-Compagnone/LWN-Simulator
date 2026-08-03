"use client";

import Image from "next/image";
import styles from "./page.module.css";
import { Button, ButtonType } from "@lwn-simulator/ui-components";

export default function Home() {
  return (
    <div className={`${styles.page} dark`}>
      <main className={styles.main}>
        <Image
          className={styles.logo}
          src="./next.svg"
          alt="Next.js logo"
          width={100}
          height={20}
          priority
        />
        <div className={styles.intro}>
          <h1>To get started, edit the page.tsx file.</h1>
          <p>
            Looking for a starting point or more instructions? Head over to{" "}
            <a
              href="https://vercel.com/templates?framework=next.js&utm_source=create-next-app&utm_medium=appdir-template-tw&utm_campaign=create-next-app"
              target="_blank"
              rel="noopener noreferrer"
            >
              Templates
            </a>{" "}
            or the{" "}
            <a
              href="https://nextjs.org/learn?utm_source=create-next-app&utm_medium=appdir-template-tw&utm_campaign=create-next-app"
              target="_blank"
              rel="noopener noreferrer"
            >
              Learning
            </a>{" "}
            center.
          </p>
        </div>
        <div className={styles.ctas}>
          <Button
            value={"Remote"}
            onClick={() => {
              window.electron.connectRemote("http://localhost:8080");
            }}
          />
          <Button
            value={"Local"}
            type={ButtonType.Secondary}
            onClick={() => {
              window.electron.connectLocal();
            }}
          />
        </div>
      </main>
    </div>
  );
}
