/** @vitest-environment jsdom */
import { describe, expect, it } from "vitest";

import { screen } from "@testing-library/react";

import { useTestI18n, useTestTheme } from "./contexts.js";
import { renderWithProviders } from "./renderWithProviders.js";

function Probe() {
  const theme = useTestTheme();
  const i18n = useTestI18n();
  return (
    <div>
      <span data-testid="theme">{theme.getCurrentPreset()?.name ?? "none"}</span>
      <span data-testid="i18n">{i18n.t("common.hello")}</span>
    </div>
  );
}

describe("renderWithProviders", () => {
  it("wraps components with theme and i18n engines", () => {
    renderWithProviders(<Probe />);
    expect(screen.getByTestId("theme").textContent).toBe("light");
    expect(screen.getByTestId("i18n").textContent).toBe("Hello");
  });
});
