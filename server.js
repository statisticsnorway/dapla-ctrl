import ViteExpress from 'vite-express';
import { createLightship } from 'lightship';
import express from 'express';
import jwt from 'jsonwebtoken';
import jwksClient from 'jwks-rsa';

import dotenv from 'dotenv';

if (!process.env.VITE_JWKS_URI) {
    dotenv.config({ path: './.env.local' });
}

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

const client = jwksClient({
    jwksUri: process.env.VITE_JWKS_URI
});

app.post('/api/verify-token', (req, res) => {
    if (!req.headers.authorization.startsWith("Bearer")) {
        return res.status(401).json({ success: false, message: 'No token provided' });
    }

    const token = req.headers.authorization.split("Bearer ")[1];

    const decodedToken = jwt.decode(token, { complete: true });
    if (!decodedToken) {
        return res.status(400).json({ success: false, message: 'Invalid token format' });
    }

    const kid = decodedToken.header.kid;
    getPublicKeyFromKeycloak(kid)
        .then(publicKey => {
            jwt.verify(token, publicKey, { algorithms: ['RS256'] }, (err, decoded) => {
                if (err) {
                    return res.status(401).json({ success: false, message: 'Invalid token' });
                }
                res.json({ success: true, user: decoded });
            });
        })
        .catch(error => {
            console.error(error);
            res.status(500).json({ success: false, message: 'Server error', error: error.message });
        });
});

app.get('/api/teams', (req, res) => {
    if (!req.headers.authorization.startsWith("Bearer")) {
        return res.status(401).json({ success: false, message: 'No token provided' });
    }

    const token = req.headers.authorization.split("Bearer ")[1];
    const url = `${process.env.VITE_DAPLA_TEAM_API_URL}/teams`;

    fetch(url, {
        method: "GET",
        headers: {
            "accept": "*/*",
            "Authorization": `Bearer ${token}`,
        }
    }).then(response => {
        if (!response.ok) {
            throw new Error('Response not ok');
        }
        return response.json();
    }).then(data => {
        return res.json({ success: true, data: data._embedded.teams });
    }).catch(error => {
        console.error(error);
        res.status(500).json({ success: false, message: 'Server error', error: error.message });
    });
});

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
