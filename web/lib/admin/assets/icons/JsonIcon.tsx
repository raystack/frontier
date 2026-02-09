import type { SVGProps } from "react";

export function JsonIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      {...props}
    >
      <path
        d="M9.99996 1.33337H3.99996C3.64634 1.33337 3.3072 1.47385 3.05715 1.7239C2.8071 1.97395 2.66663 2.31309 2.66663 2.66671V13.3334C2.66663 13.687 2.8071 14.0261 3.05715 14.2762C3.3072 14.5262 3.64634 14.6667 3.99996 14.6667H12C12.3536 14.6667 12.6927 14.5262 12.9428 14.2762C13.1928 14.0261 13.3333 13.687 13.3333 13.3334V4.66671L9.99996 1.33337Z"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M9.33337 1.33337V4.00004C9.33337 4.35366 9.47385 4.6928 9.7239 4.94285C9.97395 5.1929 10.3131 5.33337 10.6667 5.33337H13.3334"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
