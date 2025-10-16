import * as R from "ramda";
const rejectEmptyValue = R.reject(R.anyPass([R.isEmpty, R.isNil]));

const splitAndTrim = R.pipe(
  R.split(","),
  R.map(R.trim),
  R.filter(R.complement(R.isEmpty))
);

export function updateResponse(data: any) {
  data.metadata = (data.metadata || []).reduce((acc: any, metadata: any) => {
    acc[metadata.key] = metadata.value;
    return acc;
  }, {});

  const newFeatures = splitAndTrim(data.newfeatures || "");
  delete data.newfeatures;

  // Transform features back to proper format (handle both {label, value} and {name} formats)
  data.features = [...(data.features || []).map((f: any) => ({ name: f.label || f.value || f.name })), ...newFeatures.map((f) => ({ name: f }))];

  return rejectEmptyValue(data);
}
