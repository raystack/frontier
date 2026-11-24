export interface ProfileFormValues {
  title: string;
  email: string;
  avatar?: string;
}

export function profileUpdateBody(
  values: ProfileFormValues,
  currentName: string | undefined
) {
  return {
    title: values.title,
    email: values.email,
    avatar: values.avatar,
    name: currentName
  };
}
