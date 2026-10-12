// Phone page for the gym server (cmd/gymserver). Built on its own by
// vite.gym.config.js; shares the 5/3/1 math with the desktop app.
import "./gym.css";
import { mount } from "svelte";
import GymApp from "./GymApp.svelte";

export default mount(GymApp, { target: document.getElementById("app") });
