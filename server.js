import ViteExpress from 'vite-express';
import { createLightship } from 'lightship';
import express from 'express';
import jwt from 'jsonwebtoken';
import jwksClient from 'jwks-rsa';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

const client = jwksClient({
    jwksUri: 'https://keycloak.staging-bip-app.ssb.no/auth/realms/ssb/protocol/openid-connect/certs'
});

app.post('/verify-token', async (req, res) => {
    const { token } = req.body;

    try {
        const decodedToken = jwt.decode(token, { complete: true });
        const kid = decodedToken.header.kid;
        const publicKey = await getPublicKeyFromKeycloak(kid);
        jwt.verify(token, publicKey, { algorithms: ['RS256'] }, (err, decoded) => {
            if (err) {
                return res.status(401).json({ success: false, message: 'Invalid token' });
            }
            res.json({ success: true, user: decoded });
        });
    } catch (error) {
        res.status(500).json({ success: false, message: 'Server error' });
    }
});

async function getPublicKeyFromKeycloak(kid) {
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
