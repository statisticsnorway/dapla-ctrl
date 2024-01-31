<a name="readme-top"></a>

<h3 align="center">Dapla ctrl</h3>

  <p align="center">
    A web interface for performing administrative tasks related to Dapla teams.
    <br />
    <br />
    <a href="https://github.com/statisticsnorway/dapla-ctrl/issues">Report Bug</a>
    ·
    <a href="https://github.com/statisticsnorway/dapla-ctrl/issues">Request Feature</a>
  </p>
</div>

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->

## About The Project

A web interface for performing administrative tasks related to Dapla teams. TODO: Put more info here.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Built With

- [![Vite][Vite.js]][Vite-url]
- [![React][React.js]][React-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- GETTING STARTED -->

## Getting Started

This is an example of how you may give instructions on setting up your project locally.
To get a local copy up and running follow these simple example steps.

### Prerequisites

This is an example of how to list things you need to use the software and how to install them.

- npm
  ```sh
  npm install npm@latest -g
  ```
- Install nodemon (required to run the development server)
  ```
  npm install -g nodemon
  ```
- If you want to test against local version of dapla-team-api-redux. [Click here for step by step guide to set it up](https://example.com)
- Create .env.local (note you must replace dummy names with real values)
  If testing with local version of dapla-team-api-redux put this:
  ```sh
  touch .env.local && printf 'VITE_DAPLA_TEAM_API_URL="http://localhost:8080"\nVITE_JWKS_URI="https://your-keycloak.domain.com/auth/realms/ssb/protocol/openid-connect/certs"\nVITE_SSB_BEARER_URL="https://your-http-bin.domain.com/bearer"' >> .env.local
  ```
  If testing with dapla-team-api-redux in production, put this:
  ```sh
  touch .env.local && printf 'VITE_DAPLA_TEAM_API_URL="http://your-running-application.domain.com"\nVITE_JWKS_URI="https://your-keycloak.domain.com/auth/realms/ssb/protocol/openid-connect/certs"\nVITE_SSB_BEARER_URL="https://your-http-bin.domain.com/bearer"' >> .env.local
  ```

### Installation

1. Clone the repo
   ```sh
   git clone https://github.com/statisticsnorway/dapla-ctrl.git
   ```
2. Navigate into the repository
   ```sh
   cd dapla-ctrl
   ```
3. Install NPM packages
   ```sh
   npm install
   ```
4. Start the development server and access the application at http://localhost:3000
   ```sh
   npm run dev
   ```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTRIBUTING -->

## Contributing

Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- LICENSE -->

## License

Distributed under the MIT License. See `LICENSE.txt` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->

[contributors-shield]: https://img.shields.io/github/contributors/github_username/repo_name.svg?style=for-the-badge
[contributors-url]: https://github.com/github_username/repo_name/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/github_username/repo_name.svg?style=for-the-badge
[forks-url]: https://github.com/github_username/repo_name/network/members
[stars-shield]: https://img.shields.io/github/stars/github_username/repo_name.svg?style=for-the-badge
[stars-url]: https://github.com/github_username/repo_name/stargazers
[issues-shield]: https://img.shields.io/github/issues/github_username/repo_name.svg?style=for-the-badge
[issues-url]: https://github.com/github_username/repo_name/issues
[Vite.js]: https://avatars.githubusercontent.com/u/65625612?s=48&v=4
[Vite-url]: https://vitejs.dev/
[React.js]: https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB
[React-url]: https://reactjs.org/
