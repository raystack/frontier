export async function update(
  url: string,
  { arg }: { arg: Record<string, string> }
) {
  await fetch(url, {
    method: "POST",
    headers: {
      "X-Frontier-Email": "admin@raystack.org",
    },
    body: JSON.stringify(arg),
  });
}

export async function updateOrganisation(
  url: string,
  { arg }: { arg: Record<string, string> }
) {
  await fetch(url, {
    method: "POST",
    headers: {
      "X-Frontier-Email": "admin@raystack.org",
    },
    body: JSON.stringify(arg),
  });
}
