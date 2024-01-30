import ViteExpress from 'vite-express';
import { createLightship } from 'lightship';
import cache from 'memory-cache';
import express from 'express';
import jwt from 'jsonwebtoken';
import jwksClient from 'jwks-rsa';
import { getReasonPhrase } from 'http-status-codes';
import dotenv from 'dotenv';

if (!process.env.VITE_JWKS_URI) {
    dotenv.config({ path: './.env.local' });
}

const DAPLA_TEAM_API_URL = process.env.VITE_DAPLA_TEAM_API_URL;

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

const client = jwksClient({
    jwksUri: process.env.VITE_JWKS_URI
});

// Middleware to protect APi endpoints, requiring Bearer token every time.
async function tokenVerificationMiddleware(req, res, next) {
    try {
        if (!req.headers.authorization || !req.headers.authorization.startsWith("Bearer")) {
            return res.status(401).json({ message: 'No token provided' });
        }

        const token = req.headers.authorization.split("Bearer ")[1];
        const decodedToken = jwt.decode(token, { complete: true });
        if (!decodedToken) {
            return res.status(400).json({ message: 'Invalid token format' });
        }

        const kid = decodedToken.header.kid;
        const publicKey = await getPublicKeyFromKeycloak(kid);
        jwt.verify(token, publicKey, { algorithms: ['RS256'] }, (err, decoded) => {
            if (err) {
                return res.status(401).json({ message: 'Invalid token' });
            }
            req.user = decoded;
            req.token = token;
            next();
        });
    } catch (error) {
        console.error(error);
        res.status(500).json({ message: 'Server error', error: error.message });
    }
}

app.post('/api/verify-token', (req, res) => {
    if (!req.headers.authorization.startsWith("Bearer")) {
        return res.status(401).json({ message: 'No token provided' });
    }

    const token = req.headers.authorization.split("Bearer ")[1];

    const decodedToken = jwt.decode(token, { complete: true });
    if (!decodedToken) return res.status(400).json({ message: 'Invalid token format' });


    const kid = decodedToken.header.kid;
    getPublicKeyFromKeycloak(kid)
        .then(publicKey => {
            jwt.verify(token, publicKey, { algorithms: ['RS256'] }, (err, decoded) => {
                if (err) {
                    return res.status(401).json({ message: 'Invalid token' });
                }
                res.json({ user: decoded });
            });
        })
        .catch(error => {
            console.error(error);
            res.status(500).json({ message: 'Server error', error: error.message });
        });
});

app.get('/api/teamOverview', tokenVerificationMiddleware, async (req, res, next) => {
    const token = req.token;
    const principalName = req.user.email;
    const allteamsUrl = `${DAPLA_TEAM_API_URL}/teams`;
    const myTeamsUrl = `${DAPLA_TEAM_API_URL}/users/${principalName}/teams`;

    try {
        const [allTeams, myTeams] = await Promise.all([
            fetchAPIData(token, allteamsUrl, 'Failed to fetch all teams')
                .then(teams => getTeamOverviewTeams(token, teams)),
            fetchAPIData(token, myTeamsUrl, 'Failed to fetch my teams')
                .then(teams => getTeamOverviewTeams(token, teams))
        ])

        const result = {
            allTeams: {
                count: allTeams.count,
                ...allTeams._embedded
            },
            myTeams: {
                count: myTeams.count,
                ...myTeams._embedded
            }
        };

        res.json(result);
    } catch (error) {
        next(error)
    }
});

async function getTeamOverviewTeams(token, teams) {
    const teamPromises = teams._embedded.teams.map(async (team) => {
        const teamUniformName = team.uniform_name;
        const teamInfoUrl = `${DAPLA_TEAM_API_URL}/teams/${teamUniformName}`;
        const teamUsersUrl = `${DAPLA_TEAM_API_URL}/teams/${teamUniformName}/users`;
        const teamManagerUrl = `${DAPLA_TEAM_API_URL}/groups/${teamUniformName}-managers/users`;

        const [teamInfo, teamUsers, teamManager] = await Promise.all([
            fetchAPIData(token, teamInfoUrl, 'Failed to fetch team info').catch(() => {
                return {
                    uniform_name: teamUniformName,
                    section_name: "Mangler seksjon"
                }
            }),
            fetchAPIData(token, teamUsersUrl, 'Failed to fetch team users'),
            fetchAPIData(token, teamManagerUrl, 'Failed to fetch team manager')
        ]);
        team['section_name'] = teamInfo.section_name;
        team["team_user_count"] = teamUsers.count;
        team["manager"] = teamManager.count > 0 ? teamManager._embedded.users[0] : {
            "display_name": "Mangler ansvarlig",
            "principal_name": "Mangler ansvarlig",
        };

        return { ...team };
    });

    const resolvedTeams = await Promise.all(teamPromises);
    const validTeams = resolvedTeams.filter(team => team !== null);

    teams._embedded.teams = validTeams;
    teams.count = validTeams.length;
    return teams;
}

app.get('/api/userProfile/:principalName', tokenVerificationMiddleware, async (req, res, next) => {
    try {
        const token = req.token;
        const principalName = req.params.principalName;
        const userProfileUrl = `${DAPLA_TEAM_API_URL}/users/${principalName}`;
        const userManagerUrl = `${DAPLA_TEAM_API_URL}/users/${principalName}/manager`;
        const userPhotoUrl = `${DAPLA_TEAM_API_URL}/users/${principalName}/photo`;

        const cacheKey = `userProfile-${principalName}`;
        const cachedUserProfile = cache.get(cacheKey);
        if (cachedUserProfile) {
            return res.json(cachedUserProfile);
        }

        const [userProfile, userManager, userPhoto] = await Promise.all([
            fetchAPIData(token, userProfileUrl, 'Failed to fetch userProfile'),
            fetchAPIData(token, userManagerUrl, "Failed to fetch user manager"),
            fetchPhoto(token, userPhotoUrl, "Failed to fetch user photo")
        ])

        const data = { ...userProfile, manager: { ...userManager }, photo: userPhoto };
        cache.put(cacheKey, data, 3600000);

        return res.json(data);
    } catch (error) {
        next(error)
    }
});

async function fetchPhoto(token, url, fallbackErrorMessage) {
    const response = await fetch(url, getFetchOptions(token));

    if (!response.ok) {
        throw new Error(fallbackErrorMessage);
    }

    const arrayBuffer = await response.arrayBuffer();
    const photoBuffer = Buffer.from(arrayBuffer);
    return photoBuffer.toString('base64');
}

async function getUserProfileTeamData(token, principalName, teams) {
    const teamPromises = teams._embedded.teams.map(async (team) => {
        const teamUniformName = team.uniform_name;
        const teamInfoUrl = `${DAPLA_TEAM_API_URL}/teams/${teamUniformName}`;
        const teamGroupsUrl = `${DAPLA_TEAM_API_URL}/teams/${teamUniformName}/groups`;
        const teamManagerUrl = `${DAPLA_TEAM_API_URL}/groups/${teamUniformName}-managers/users`;

        const [teamInfo, teamGroups, teamManager] = await Promise.all([
            fetchAPIData(token, teamInfoUrl, 'Failed to fetch team info').catch(() => {
                return {
                    uniform_name: teamUniformName,
                    section_name: "Mangler seksjon"
                }
            }),
            fetchAPIData(token, teamGroupsUrl, 'Failed to fetch groups').then(response => {
                const groupPromises = response._embedded.groups.map(group => fetchUserGroups(group, token, principalName));
                return Promise.all(groupPromises).then(groupsArrays => groupsArrays.flat());
            }),
            fetchAPIData(token, teamManagerUrl, 'Failed to fetch team manager')
        ]);

        team['section_name'] = teamInfo.section_name;
        team["manager"] = teamManager.count > 0 ? teamManager._embedded.users[0] : {
            "display_name": "Mangler ansvarlig",
            "principal_name": "Mangler ansvarlig",
        };
        team["groups"] = teamGroups;

        return { ...team };
    });

    const resolvedTeams = await Promise.all(teamPromises);
    const validTeams = resolvedTeams.filter(team => team !== null);

    teams._embedded.teams = validTeams;
    teams.count = validTeams.length;
    return teams;
}

async function fetchUserGroups(group, token, principalName) {
    const groupUsersUrl = `${DAPLA_TEAM_API_URL}/groups/${group.uniform_name}/users`;
    try {
        const groupUsers = await fetchAPIData(token, groupUsersUrl, 'Failed to fetch group users');
        if (!groupUsers._embedded || !groupUsers._embedded.users || groupUsers._embedded.users.length === 0) {
            return [];
        }

        return groupUsers._embedded.users
            .filter(user => user.principal_name === principalName)
            .map(() => group.uniform_name);
    } catch (error) {
        console.error(`Error processing group ${group.uniform_name}:`, error);
        throw error;
    }
}

app.get('/api/userProfile/:principalName/team', tokenVerificationMiddleware, async (req, res, next) => {
    const token = req.token;
    const principalName = req.params.principalName;
    const myTeamsUrl = `${DAPLA_TEAM_API_URL}/users/${principalName}/teams`;

    try {
        const [myTeams] = await Promise.all([
            fetchAPIData(token, myTeamsUrl, 'Failed to fetch my teams')
                .then(teams => getUserProfileTeamData(token, principalName, teams))
        ])

        const result = {
            count: myTeams.count,
            ...myTeams._embedded
        };

        res.json(result);
    } catch (error) {
        next(error)
    }
});

async function fetchAPIData(token, url, fallbackErrorMessage) {
    const response = await fetch(url, getFetchOptions(token));
    const wwwAuthenticate = response.headers.get('www-authenticate');

    if (!response.ok) {
        const { error_description } = wwwAuthenticate ? parseWwwAuthenticate(wwwAuthenticate) : { error_description: fallbackErrorMessage };
        throw new APIError(error_description, response.status);
    }

    return response.json();
}

function parseWwwAuthenticate(header) {
    const parts = header.split(',');
    const result = {};

    parts.forEach(part => {
        const [key, value] = part.trim().split('=');
        result[key] = value.replace(/"/g, '');
    });

    return result;
}

class APIError extends Error {
    constructor(message, statusCode) {
        super(message);
        this.statusCode = statusCode;
    }
}

function getFetchOptions(token) {
    return {
        method: "GET",
        headers: {
            "accept": "*/*",
            "Authorization": `Bearer ${token}`,
        }
    };
}

function getPublicKeyFromKeycloak(kid) {
    return new Promise((resolve, reject) => {
        client.getSigningKey(kid, (err, key) => {
            if (err) {
                reject(err);
                return;
            }
            if (!key) {
                reject(new Error('No key found'));
                return;
            }
            resolve(key.getPublicKey());
        });
    });
}

app.use((err, req, res, next) => {
    const statusCode = err.statusCode || 500;

    return res.status(statusCode).json({
        success: false,
        error: {
            code: getReasonPhrase(statusCode),
            message: err.message
        }
    });
});

//const lightship = await createLightship();
// Replace above with below to get liveness and readiness probes when running locally
const lightship = await createLightship({ detectKubernetes: false });

ViteExpress.listen(app, PORT, () => {
    lightship.signalReady();
    console.log(
        `Server is listening on port ${PORT} ... ${process.env.NODE_ENV}`
    );
}).on('error', () => {
    lightship.shutdown();
});

lightship.registerShutdownHandler(async () => {
    console.log('Server is shutting down...');
});
