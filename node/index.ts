import fs from 'fs';
import os from 'os';
import crypto from 'crypto';
import dotenv from 'dotenv';

import { CID } from 'multiformats/cid';
import { sha256 } from 'multiformats/hashes/sha2';


import { FileStreamFactory, TurboFactory } from '@ardrive/turbo-sdk/node';
import Arweave from 'arweave';
import { JWKInterface } from 'arbundles/node';
import path from 'path';
import { ExecException, exec } from 'child_process';

import { fileURLToPath } from 'url';
import { dirname } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
import express from 'express';

const app = express();
const port = 3000;

dotenv.config();


process.on('uncaughtException', (error) => {
    console.error('Uncaught Exception:', error);
});

process.on('unhandledRejection', (reason, promise) => {
    console.error('Unhandled Rejection at:', promise, 'reason:', reason);
});


const POSTGRES_DB = process.env.POSTGRES_DB;
const POSTGRES_HOST = process.env.POSTGRES_HOST;
const POSTGRES_USER = process.env.POSTGRES_USER;
const POSTGRES_PASSWORD = process.env.POSTGRES_PASSWORD;
const POSTGRES_PORT = process.env.POSTGRES_PORT;



app.use(express.json()); // To parse JSON bodies

// Define a type for the rejection reason
type ExecPromiseError = {
    error: ExecException | null;
    stderr: string;
};

// Wrap exec in a promise to use it with async/await
const execPromise = (command: string, options = {}) => {
    return new Promise<{ stdout: string; stderr: string }>((resolve, reject) => {


        exec(command, options, (error: ExecException | null, stdout: string, stderr: string) => {
            if (error) {
                reject({ error: error, stderr: stderr }) // Reject the promise if there's an error
            } else {
                resolve({ stdout, stderr }) // Resolve the promise if there's no error
            }
        })
    })
}






/*
// load your JWK from a file or generate a new one
const arweave = Arweave.init({
    host: 'arweave.net',
    port: 443,
    protocol: 'https',

});
*/
const jwk: JWKInterface = JSON.parse(fs.readFileSync('./arweave-wallet.json').toString());



const turbo = TurboFactory.authenticated({ privateKey: jwk as unknown as JWKInterface });

app.post('/arweave/upload/string', async (req, res) => {
    const inputString = req.body.cid
    console.log(inputString)

    const dbName = POSTGRES_DB
    const dbUser = POSTGRES_USER
    const dbHost = POSTGRES_HOST
    const dbPassword = POSTGRES_PASSWORD
    const dbPort = POSTGRES_PORT

    const env = Object.create(process.env);
    env.PGPASSWORD = dbPassword;

    const command = `pg_dump -U ${dbUser} -h ${dbHost} -p ${dbPort} ${dbName}`;
    console.log("Dumping db...")

    const options = { env, maxBuffer: 1000 * 1024 * 1024 }; // 1000 MB

    try {
        const { stdout, stderr } = await execPromise(command, options);

        // Encrypt using AES and upload to Arweave
        const ENCRYPTION_KEY = process.env.DB_KEY;
        if (!ENCRYPTION_KEY) {
            throw new Error("DB_KEY not set");
        }

        // AES encryption
        const iv = crypto.randomBytes(16); // Generate arandom IV
        const cipher = crypto.createCipheriv('aes-256-cbc', Buffer.from(ENCRYPTION_KEY, 'base64'), iv);
        let encrypted = cipher.update(stdout, 'utf8', 'hex');
        encrypted += cipher.final('hex');
        const encryptedOutput = iv.toString('hex') + ':' + encrypted; // Prepend the IV for decryption purposes

        // Generate CIDv1 for the encrypted output
        const contentBytes = Buffer.from(encryptedOutput, 'utf8');
        const hash = await sha256.digest(contentBytes);
        const cid = CID.create(1, 0x55, hash) // 0x55 is the code for raw binary data

        const timestamp = Date.now();
        console.log(`DB dump and encryption successful at ${timestamp}`);
        console.log(`CIDv1: ${cid.toString()}`);


        const encryptedFilename = `encrypted_db_dump_${timestamp}.txt`;
        const cidFilename = `cid_${Date.now()}.txt`;
        const cidFilePath = path.join(os.tmpdir(), cidFilename);
        const filePath = path.join(__dirname, encryptedFilename);


        try {
            // Save encrypted output to a file
            fs.writeFileSync(filePath, encryptedOutput);
            fs.writeFileSync(cidFilePath, cid.toString());

            console.log(`Encrypted DB dump saved to ${filePath}`);

            // Preparing for upload to Arweave
            const bufferSize = fs.statSync(cidFilePath).size;
            const bufferSizeFactory = () => bufferSize;
            const bufferStreamFactory: FileStreamFactory = () => fs.createReadStream(cidFilePath);

            // Get the wallet balance
            const { winc: balance } = await turbo.getBalance();
            // Get the cost of uploading the file
            const [{ winc: bufferSizeCost }] = await turbo.getUploadCosts({
                bytes: [bufferSize],
            });

            // check if balance greater than upload cost
            if (parseInt(balance) < parseInt(bufferSizeCost)) {
                return res.status(400).send("Insufficient balance to upload string!\nYour balance: " + balance + "\nCost to upload: " + bufferSizeCost + "\nPlease top up your wallet and try again.")
            }


            // Upload the data
            const { id, owner } = await turbo.uploadFile({ fileStreamFactory: bufferStreamFactory, fileSizeFactory: bufferSizeFactory })

            console.log(`File uploaded successfully with ID: ${id}, Owner: ${owner}`);
            fs.unlinkSync(cidFilePath);

            res.json({
                message: cid.toString(),
                transactionId: id,
                owner: owner,
            });


        } catch (error) {
            console.error("Failed to upload encrypted database dump!", error);
            res.status(500).send("Failed to upload encrypted database dump to Arweave!");
            // Clean up by deleting the local encrypted file
            fs.unlinkSync(filePath);
            if (cidFilePath && fs.existsSync(cidFilePath)) {
                fs.unlinkSync(cidFilePath);
            }
        }

    } catch (err) {
        const { error, stderr } = err as ExecPromiseError;
        console.error(`Getting database failed: ${error}`)
        res.status(500).send('Backup failed: ' + stderr);
    }
})

app.listen(port, () => {
    console.log(`Server running at http://localhost:${port}`);
});