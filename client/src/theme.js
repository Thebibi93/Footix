import { createTheme } from "@mantine/core";

export const theme = createTheme({
  primaryColor: "cyan",
  defaultRadius: "xl",
  fontFamily:
    "Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
  headings: {
    fontFamily:
      "Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
    fontWeight: "800",
  },
  colors: {
    stadium: [
      "#e6fcf5",
      "#c3fae8",
      "#96f2d7",
      "#63e6be",
      "#38d9a9",
      "#20c997",
      "#12b886",
      "#0ca678",
      "#099268",
      "#087f5b",
    ],
  },
  components: {
    Paper: {
      defaultProps: {
        radius: "xl",
        withBorder: true,
        shadow: "sm",
      },
    },
    Card: {
      defaultProps: {
        radius: "xl",
        withBorder: true,
        shadow: "sm",
        padding: "lg",
      },
    },
    Button: {
      defaultProps: {
        radius: "xl",
      },
    },
    TextInput: {
      defaultProps: {
        radius: "xl",
      },
    },
    PasswordInput: {
      defaultProps: {
        radius: "xl",
      },
    },
    SegmentedControl: {
      defaultProps: {
        radius: "xl",
      },
    },
    Tabs: {
      defaultProps: {
        radius: "xl",
      },
    },
    Badge: {
      defaultProps: {
        radius: "xl",
      },
    },
  },
});
