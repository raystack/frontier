import type { SVGProps } from "react";

export function KeyIcon(props: SVGProps<SVGSVGElement>) {
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
        d="M10.5 3.5C11.6046 3.5 12.5 4.39543 12.5 5.5M14.5 5.5C14.5 7.70914 12.7091 9.5 10.5 9.5C10.2662 9.5 10.037 9.47994 9.8142 9.44144C9.43885 9.37658 9.04134 9.45866 8.772 9.728L7 11.5H5.5V13H4V14.5H1.5V12.6213C1.5 12.2235 1.65804 11.842 1.93934 11.5607L6.272 7.228C6.54134 6.95866 6.62342 6.56115 6.55856 6.1858C6.52006 5.96297 6.5 5.73383 6.5 5.5C6.5 3.29086 8.29086 1.5 10.5 1.5C12.7091 1.5 14.5 3.29086 14.5 5.5Z"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
