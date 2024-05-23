/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./src/**/*.{html,templ}"],
  theme: {
    extend: {
      fontFamily: {
        major: ["Major Mono Display", "monospace"],
      },
    },
  },
};
