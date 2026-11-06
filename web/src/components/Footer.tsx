import type { Meta } from "../types";

interface FooterProps {
  meta: Meta | null;
}

/** Floating footer showing the running build version, centered at the bottom of the viewport. */
export function Footer({ meta }: FooterProps) {
  if (!meta) return null;

  return <footer id="build-footer">Skygate {meta.buildVersion}</footer>;
}
