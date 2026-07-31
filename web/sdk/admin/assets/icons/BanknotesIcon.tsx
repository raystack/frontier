import type { SVGProps } from "react";

/*
 * Heroicons "banknotes" (24px outline), redrawn on the 16 grid the other icons use.
 * - Stroke is 1, not the 1.5 of the upstream icon: stroke width is relative to
 *   the viewBox, so scaling 24 -> 16 has to scale the stroke too. Leaving it at
 *   1.5 renders 1.5x too heavy, which shows badly at the 68px empty-state size.
 */
export function BanknotesIcon(props: SVGProps<SVGSVGElement>) {
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
        d="M1.5 12.5C5.1448 12.5 8.6757 12.9875 12.0312 13.9008C12.5159 14.0327 13 13.6724 13 13.1701V12.5M2.5 3V3.5C2.5 3.7761 2.2761 4 2 4H1.5M1.5 4V3.75C1.5 3.3358 1.8358 3 2.25 3H13.5M1.5 4V10M13.5 3V3.5C13.5 3.7761 13.7239 4 14 4H14.5M13.5 3H13.75C14.1642 3 14.5 3.3358 14.5 3.75V10.25C14.5 10.6642 14.1642 11 13.75 11H13.5M14.5 10H14C13.7239 10 13.5 10.2239 13.5 10.5V11M13.5 11H2.5M2.5 11H2.25C1.8358 11 1.5 10.6642 1.5 10.25V10M2.5 11V10.5C2.5 10.2239 2.2761 10 2 10H1.5M10 7C10 8.1046 9.1046 9 8 9C6.8954 9 6 8.1046 6 7C6 5.8954 6.8954 5 8 5C9.1046 5 10 5.8954 10 7ZM12 7H12.005V7.005H12V7ZM4 7H4.005V7.005H4V7Z"
        stroke="currentColor"
        strokeWidth="1"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
