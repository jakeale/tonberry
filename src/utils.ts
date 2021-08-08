/**
 * Check if a given object is empty.
 * @param obj
 * @returns True if empty, else False
 */
export const isEmpty = (obj: object) => {
  return Object.keys(obj).length === 0;
};
