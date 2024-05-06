# Introduction

Frontier SDK provides wrapper arround [Frontier Apis](/apis/frontier-administration-api) and React components which you can integrate in your web app for user and organization management.

## Setup

Install the sdk in your project.

```sh
npm i @raystack/frontier
```

Import the `FrontierProvider` in the root of your project and wrap your application code with it.

```javascript
import { FrontierProvider } from "@raystack/frontier/react";

const frontierConfig = {};

function App() {
  return (
    <FrontierProvider config={frontierConfig}>
      /* Your app code here */
    </FrontierProvider>
  );
}
```

To access the frontier instance inside your application, import the `useFrontier` hook in your components.

```javascript
import { useFrontier } from "@raystack/frontier/react";

export function DemoComponent() {
  const { client } = useFrontier();
  const [userLoader, setUserLoader] = useState(true);
  const [user, setUser] = useState();

  useEffect(() => {
    async function getCurrentUser() {
      try {
        setUserLoader(true);
        const userResp = await client.frontierServiceGetCurrentUser();
        if (userResp?.data) {
          setUseruser(userResp?.data?.user);
        }
      } catch (err) {
        console.error(err);
      } finally {
        setUserLoader(false);
      }
    }
    getCurrentUser();
  }, [client]);

  return <div>demo</div>;
}
```
