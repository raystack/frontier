# @raystack/frontier

Frontier JS SDK allows you to implement authentication in your [React](https://reactjs.org/) application quickly using Magic links and social sign-in. It also allows you to access the
SignIn, SignUp, user profile, Workspace creation, Workspace Profile etc. components.

Here is a quick guide on getting started with `@raystack/frontier` package.

## Step 1 - Create Instance

Get Frontier URL by instantiating [Frontier instance](https://github.com/odpf/frontier)
.

## Step 2 - Install package

Install `@raystack/frontier` library

```sh
npm i --save @raystack/frontier
OR
yarn add @raystack/frontier
```

## Step 3 - Configure Provider and use Frontier Components

Frontier comes with [react context](https://reactjs.org/docs/context.html) which serves as `Provider` component for the application

```jsx
import {
  FrontierProvider,
  Frontier,
  useFrontier
} from '@raystack/frontier/react';

const App = () => {
  return (
    <FrontierProvider
      config={{
        endpoint: 'http://localhost:3000',
        redirectLogin: window.location.origin,
        redirectSignup: window.location.origin,
        redirectMagicLinkVerify: window.location.origin
      }}
    >
      <SignIn />
    </FrontierProvider>
  );
};

const Profile = () => {
  const { user } = useFrontier();
  if (user) {
    return <div>{user.email}</div>;
  }
  return null;
};
```
