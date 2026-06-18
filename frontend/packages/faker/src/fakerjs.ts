import { faker } from "@faker-js/faker";
import type { FakerPort } from "./port.js";

/** @faker-js/faker adapter (rich locale data). */
export function createFakerJSAdapter(seed?: number): FakerPort {
  if (seed !== undefined) {
    faker.seed(seed);
  }

  return {
    uuid: () => faker.string.uuid(),
    email: () => faker.internet.email(),
    firstName: () => faker.person.firstName(),
    lastName: () => faker.person.lastName(),
    companyName: () => faker.company.name(),
    sentence: (wordCount = 6) => faker.lorem.sentence(wordCount),
    int: (min = 0, max = 100) => faker.number.int({ min, max }),
  };
}
