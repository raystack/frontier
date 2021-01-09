import * as R from 'ramda';

export const getEmailFromIAPHeader = (header: string) => {
  return header?.replace('accounts.google.com:', '');
};

export const getUsernameFromEmail = (email: string) => {
  return email?.split('@')?.shift();
};

export const getUsernameFromIAPHeader = R.pipe(
  getEmailFromIAPHeader,
  getUsernameFromEmail
);
