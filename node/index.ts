import fs from 'fs';
import os from 'os';

import bodyParser from 'body-parser';
import { CurrencyMap, FileStreamFactory, TurboFactory } from '@ardrive/turbo-sdk/node';
import Arweave from 'arweave';
import { JWKInterface } from 'arbundles/node';
import path from 'path';

import { fileURLToPath } from 'url';
import { dirname } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
import express from 'express';

const app = express();
const port = 3000;


app.use(bodyParser.text())




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
    const inputString = req.body;
    console.log(inputString)


    // Create a temporary file
    const tempFilePath = path.join(os.tmpdir(), `temp-${Date.now()}.txt`);






    // Get the cost of uploading the file
    try {
        fs.writeFileSync(tempFilePath, inputString)
        // Get the wallet balance
        const { winc: balance } = await turbo.getBalance();

        // Prepare the data for upload
        const bufferSize = fs.statSync(tempFilePath).size;




        // Get the cost of uploading the file
        const [{ winc: bufferSizeCost }] = await turbo.getUploadCosts({
            bytes: [bufferSize],
        });

        // check if balance greater than upload cost
        if (parseInt(balance) < parseInt(bufferSizeCost)) {
            return res.status(400).send("Insufficient balance to upload string!\nYour balance: " + balance + "\nCost to upload: " + bufferSizeCost + "\nPlease top up your wallet and try again.")
        }

        // Upload the data
        const bufferSizeFactory = () => bufferSize;
        const bufferStreamFactory: FileStreamFactory = () => {
            const stream = fs.createReadStream(tempFilePath);
            return stream;
        };
        const { id, owner } = await turbo.uploadFile({ fileStreamFactory: bufferStreamFactory, fileSizeFactory: bufferSizeFactory });

        fs.unlinkSync(tempFilePath)

        return res.json({
            message: inputString,
            transactionId: id,
            owner,
        })
    } catch (error) {
        console.error("Failed to upload data item!", error)

        try { fs.unlinkSync(tempFilePath); } catch (error) { console.error("Failed to delete temp file!", error) }

        return res.status(500).send("Failed to upload string to Arweave!")
    }


})

app.listen(port, () => {
    console.log(`Server running at http://localhost:${port}`);
});