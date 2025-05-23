# OBSEA Digital Twin Frontend

This repository contains the frontend application for the OBSEA Digital Twin project. Built with Vue.js and TypeScript, it provides an interactive web interface for visualizing oceanographic data from the OBSEA underwater observatory.

## Technologies Used

- **Vue.js (with TypeScript):** Progressive JavaScript framework for building the user interface.
- **Mapbox GL JS:** JavaScript library for rendering interactive maps.
- **Tailwind CSS:** Utility-first CSS framework for styling.

## Project Setup

To get the frontend development environment up and running on your local machine, follow these steps:

### Prerequisites

Ensure you have Node.js and npm (or yarn) installed. You can download them from [https://nodejs.org/](https://nodejs.org/).

### Installation

1.  Clone the repository if you haven't already:

    ```bash
    git clone <repository_url>
    cd <repository_directory>/frontend
    ```

2.  Install the project dependencies:

    ```sh
    npm install
    ```

    or if you use yarn:

    ```sh
    yarn install
    ```

### Configuration (.env File)

The frontend requires a Mapbox Access Token to load and display the base map and data layers.

1.  Create a `.env` file in the root of the `frontend/` directory (where this README is located).
2.  Add the following variable to the `.env` file, replacing `YOUR_MAPBOX_ACCESS_TOKEN` with your actual token:

    ```env
    VITE_MAPBOX_ACCESS_TOKEN=YOUR_MAPBOX_ACCESS_TOKEN
    ```

**Where to get a Mapbox Access Token:**

You can obtain a free Mapbox Access Token by signing up for an account on the Mapbox website: [https://account.mapbox.com/](https://account.mapbox.com/). Once logged in, you can create and manage your access tokens. It is recommended to use a public token with appropriate restrictions for web applications.

### Running the Project

#### Compile and Hot-Reload for Development

To run the application in development mode with hot-reloading, use the following command:

```sh
npm run dev
```

or if you use yarn:

```sh
yarn dev
```

This will start a local development server, and the application will typically be accessible at `http://localhost:5173/` (or another port indicated in your terminal). Changes to the code will automatically trigger a browser refresh.

#### Type-Check, Compile and Minify for Production

To build the application for production deployment, use the following command:

```sh
npm run build
```

or if you use yarn:

```sh
yarn build
```

This command will type-check your code, compile it, and minify the output into the `dist/` directory. The files in the `dist/` directory are ready to be served by a web server for production use.
