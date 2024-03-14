import { DateTime } from "ts-luxon"

const humanizedTimePeriodToNow = (
  jsDateAllowedInput: string | number | Date
): string => {
  const time = DateTime.fromJSDate(new Date(jsDateAllowedInput))
  const diff = DateTime.now().diff(time, [
    "years",
    "months",
    "days",
    "hours",
    "minutes",
  ])
  const humanizedString = Object.keys(diff.toObject()).reduce((acc, key) => {
    if (diff.toObject()[key] !== 0) {
      acc.push(`${Math.round(Math.abs(diff.toObject()[key]))} ${key}`)
    }
    return acc
  }, [] as string[])

  return humanizedString.join(", ")
}

export default humanizedTimePeriodToNow
