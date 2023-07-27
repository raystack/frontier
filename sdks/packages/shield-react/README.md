# frontier-react

Frontier React SDK allows you to implement authentication in your [React](https://reactjs.org/) application quickly using Magic links and social sign-in. It also allows you to access the
SignIn, SignUp, user profile, Workspace creation, Workspace Profile etc. components.

Here is a quick guide on getting started with `@raystack/frontier-react` package.

## Step 1 - Create Instance

Get Frontier URL by instantiating [Frontier instance](https://github.com/odpf/frontier)
.

## Step 2 - Install package

Install `@raystack/frontier-react` library

```sh
npm i --save @raystack/frontier-react
OR
yarn add @raystack/frontier-react
```

## Step 3 - Configure Provider and use Frontier Components

Frontier comes with [react context](https://reactjs.org/docs/context.html) which serves as `Provider` component for the application

```jsx
import { FrontierProvider, Frontier, useFrontier } from "@raystack/frontier-react";

const App = () => {
  return (
    <FrontierProvider
      config={{
        frontierUrl: "http://localhost:8080",
        redirectURL: window.location.origin,
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

## Commands

### Local Development

The recommended workflow is to run react frontier in one terminal:

```bash
npm start # or yarn start
```

This builds to `/dist` and runs the project in watch mode so any edits you save inside `src` causes a rebuild to `/dist`.

Then run any example package which use frontier-react:

### Example package

Then run the example package inside another terminal:

```bash
cd example-package
npm i
npm start # or yarn start
```

The default example imports and live reloads whatever is in `/dist`, so if you are seeing an out of date component, make sure tsup is running in watch mode like we recommend above.

To do a one-off build, use `npm run build` or `yarn build`.
To run tests, use `npm test` or `yarn test`.

## Configuration

Code quality is set up for you with `prettier`, `husky`, and `lint-staged`. Adjust the respective fields in `package.json` accordingly.

### Jest

Jest tests are set up to run with `npm test` or `yarn test`.

### Bundle analysis

Calculates the real cost of your library using [size-limit](https://github.com/ai/size-limit) with `npm run size`
