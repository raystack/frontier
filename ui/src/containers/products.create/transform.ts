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

  data.behavior_config = Object.keys(data.behavior_config || {}).reduce(
    (acc: Record<string, number>, key: string) => {
      if (data.behavior_config[key]) {
        acc[key] = parseInt(data.behavior_config[key] || "");
      }
      return acc;
    },
    {}
  );

  const newFeatures = splitAndTrim(data.newfeatures || "");
  delete data.newfeatures;
  data.features = [...data.features, ...newFeatures.map((f) => ({ name: f }))];

  return rejectEmptyValue(data);
}
