// Copyright 2018 The go-juchain Authors
// This file is part of the go-juchain library.
//
// The go-juchain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-juchain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-juchain library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"github.com/juchain/go-juchain/common"
)

// binary code generated from vm/solc/gen_dappmanager.solc
// format then from this site: https://www.json.cn/#
var DAPPContractABI =`
[
    {
        "constant":false,
        "inputs":[
            {"name":"dappName","type":"string"},
            {"name":"orgName","type":"string"},
            {"name":"orgDescription","type":"string"},
            {"name":"nationalityCode","type":"uint8"},
            {"name":"ledgerReplicated","type":"uint8"},
            {"name":"icon","type":"bytes32"},
            {"name":"ext0","type":"string"},
            {"name":"ext1","type":"string"},
            {"name":"ext2","type":"string"}
        ],
        "name":"registerDAppInfo",
        "outputs":[

        ],
        "payable":true,
        "stateMutability":"payable",
        "type":"function"
    },
    {
        "constant":false,
        "inputs":[
            {
                "name":"orgDescription",
                "type":"string"
            },
            {
                "name":"ledgerReplicated",
                "type":"uint8"
            },
            {
                "name":"icon",
                "type":"bytes32"
            }
        ],
        "name":"updateDAppInfo",
        "outputs":[

        ],
        "payable":true,
        "stateMutability":"payable",
        "type":"function"
    },
    {
        "constant":true,
        "inputs":[
            {
                "name":"dappId",
                "type":"address"
            }
        ],
        "name":"getDAppInfo",
        "outputs":[
            {
                "name":"dappAddress",
                "type":"address"
            },
            {
                "name":"dappName",
                "type":"string"
            },
            {
                "name":"orgName",
                "type":"string"
            },
            {
                "name":"orgDescription",
                "type":"string"
            },
            {
                "name":"nationalityCode",
                "type":"uint8"
            },
            {
                "name":"ledgerReplicated",
                "type":"uint8"
            },
            {
                "name":"state",
                "type":"uint8"
            },
            {
                "name":"lastActive",
                "type":"uint256"
            },
            {
                "name":"icon",
                "type":"bytes32"
            }
        ],
        "payable":false,
        "stateMutability":"view",
        "type":"function"
    },
    {
        "inputs":[

        ],
        "payable":false,
        "stateMutability":"nonpayable",
        "type":"constructor"
    },
    {
        "anonymous":false,
        "inputs":[

        ],
        "name":"UpdateDAppInfoAccepted",
        "type":"event"
    },
    {
        "anonymous":false,
        "inputs":[

        ],
        "name":"RegisterDAppInfoAccepted",
        "type":"event"
    }
]`
var DAPPContractBinCode = "608060405234801561001057600080fd5b50610a07806100206000396000f3006080604052600436106100565763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166364073385811461005b578063ce9925c3146101f8578063d98f356b1461024f575b600080fd5b6040805160206004803580820135601f81018490048402850184019095528484526101f694369492936024939284019190819084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a99988101979196509182019450925082915084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a99988101979196509182019450925082915084018382808284375050604080516020601f60608a01358b0180359182018390048302840183018552818452989b60ff8b3581169c848d01359091169b958601359a91995097506080909401955091935091820191819084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a99988101979196509182019450925082915084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a9998810197919650918201945092508291508401838280828437509497506104129650505050505050565b005b6040805160206004803580820135601f81018490048402850184019095528484526101f69436949293602493928401919081908401838280828437509497505050833560ff1694505050602090910135905061062e565b34801561025b57600080fd5b5061027d73ffffffffffffffffffffffffffffffffffffffff600435166106c5565b6040805173ffffffffffffffffffffffffffffffffffffffff8b16815260ff808816608083015286811660a0830152851660c082015260e08101849052610100810183905261012060208083018281528c51928401929092528b5192939192918401916060850191610140860191908e019080838360005b8381101561030d5781810151838201526020016102f5565b50505050905090810190601f16801561033a5780820380516001836020036101000a031916815260200191505b5084810383528b5181528b516020918201918d019080838360005b8381101561036d578181015183820152602001610355565b50505050905090810190601f16801561039a5780820380516001836020036101000a031916815260200191505b5084810382528a5181528a516020918201918c019080838360005b838110156103cd5781810151838201526020016103b5565b50505050905090810190601f1680156103fa5780820380516001836020036101000a031916815260200191505b509c5050505050505050505050505060405180910390f35b336000908152602081905260409020600b015460ff161561043257600080fd5b604080516101a0810182523380825260208083018d81528385018d9052606084018c905260ff8b811660808601528a1660a085015260c08401899052600160e085018190526101008501899052610120850188905261014085018790524261016086015261018085018190526000938452838352949092208351815473ffffffffffffffffffffffffffffffffffffffff191673ffffffffffffffffffffffffffffffffffffffff90911617815591518051939492936104f9938501929190910190610940565b5060408201518051610515916002840191602090910190610940565b5060608201518051610531916003840191602090910190610940565b50608082015160048201805460a085015160ff1991821660ff9485161761ff00191661010091851682021790925560c0850151600585015560e08501516006850180549092169316929092179091558201518051610599916007840191602090910190610940565b5061012082015180516105b6916008840191602090910190610940565b5061014082015180516105d3916009840191602090910190610940565b50610160820151600a82015561018090910151600b909101805460ff19169115159190911790556040517f6142ff54c11c1eb59d0a251a917b38ccd85c84dc4c83c021e8662a73d557b00590600090a1505050505050505050565b336000908152602081905260408120600b015460ff16151560011461065257600080fd5b503360009081526020818152604090912084519091610678916003840191870190610940565b5060048101805461ff00191661010060ff861602179055600581018290556040517f3c22cfbecb928778078fb52ac2ece267e23d58f14308b86f1645e39aaad0197890600090a150505050565b73ffffffffffffffffffffffffffffffffffffffff81166000908152602081905260408120600b01546060908190819084908190819081908190819060ff16151560011461071257600080fd5b5073ffffffffffffffffffffffffffffffffffffffff8a811660009081526020818152604091829020805460048201546006830154600a8401546005850154600180870180548a51600261010094831615850260001901909216829004601f81018c90048c0282018c01909c528b8152989b97909716999098968b019760038c019760ff80891698949094048416969390931694939290918a918301828280156107fd5780601f106107d2576101008083540402835291602001916107fd565b820191906000526020600020905b8154815290600101906020018083116107e057829003601f168201915b50508a5460408051602060026001851615610100026000190190941693909304601f8101849004840282018401909252818152959d508c94509250840190508282801561088b5780601f106108605761010080835404028352916020019161088b565b820191906000526020600020905b81548152906001019060200180831161086e57829003601f168201915b5050895460408051602060026001851615610100026000190190941693909304601f8101849004840282018401909252818152959c508b9450925084019050828280156109195780601f106108ee57610100808354040283529160200191610919565b820191906000526020600020905b8154815290600101906020018083116108fc57829003601f168201915b50505050509550995099509950995099509950995099509950509193959799909294969850565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061098157805160ff19168380011785556109ae565b828001600101855582156109ae579182015b828111156109ae578251825591602001919060010190610993565b506109ba9291506109be565b5090565b6109d891905b808211156109ba57600081556001016109c4565b905600a165627a7a723058208c10352e639500b0bc4bcdf1027b0cdf5896e3b0cd3cd90144d0ff71ff31c5020029";
var DAPPContractAddress common.Address;
var DPOSBallotABI = `
[
    {
        "constant":true,
        "inputs":[
            {
                "name":"nodeid",
                "type":"string"
            }
        ],
        "name":"delegatorInfo",
        "outputs":[
            {
                "name":"ip",
                "type":"string"
            },
            {
                "name":"port",
                "type":"uint256"
            },
            {
                "name":"ticket",
                "type":"uint256"
            }
        ],
        "payable":false,
        "stateMutability":"view",
        "type":"function"
    },
    {
        "constant":true,
        "inputs":[

        ],
        "name":"delegatorList",
        "outputs":[
            {
                "name":"result",
                "type":"string"
            }
        ],
        "payable":false,
        "stateMutability":"view",
        "type":"function"
    },
    {
        "constant":true,
        "inputs":[

        ],
        "name":"birusu",
        "outputs":[
            {
                "name":"",
                "type":"address"
            }
        ],
        "payable":false,
        "stateMutability":"view",
        "type":"function"
    },
    {
        "constant":false,
        "inputs":[
            {
                "name":"nodeid",
                "type":"string"
            },
            {
                "name":"ip",
                "type":"string"
            },
            {
                "name":"port",
                "type":"uint256"
            }
        ],
        "name":"register",
        "outputs":[

        ],
        "payable":true,
        "stateMutability":"payable",
        "type":"function"
    },
    {
        "constant":false,
        "inputs":[
            {
                "name":"nodeid",
                "type":"string"
            }
        ],
        "name":"vote",
        "outputs":[

        ],
        "payable":true,
        "stateMutability":"payable",
        "type":"function"
    },
    {
        "inputs":[

        ],
        "payable":false,
        "stateMutability":"nonpayable",
        "type":"constructor"
    },
    {
        "anonymous":false,
        "inputs":[

        ],
        "name":"NotifyRegistered",
        "type":"event"
    },
    {
        "anonymous":false,
        "inputs":[

        ],
        "name":"NotifyVoted",
        "type":"event"
    }
]`
var DPOSBallotBinCode = "608060405234801561001057600080fd5b50610d28806100206000396000f300608060405260043610610057576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063640733851461005c578063ce9925c31461023e578063d98f356b146102b5575b600080fd5b61023c600480360381019080803590602001908201803590602001908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509192919290803590602001908201803590602001908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509192919290803590602001908201803590602001908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509192919290803560ff169060200190929190803560ff1690602001909291908035600019169060200190929190803590602001908201803590602001908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509192919290803590602001908201803590602001908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509192919290803590602001908201803590602001908080601f01602080910402602001604051908101604052809392919081815260200183838082843782019150505050505091929192905050506104b9565b005b6102b3600480360381019080803590602001908201803590602001908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509192919290803560ff16906020019092919080356000191690602001909291905050506107a9565b005b3480156102c157600080fd5b506102f6600480360381019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506108c1565b604051808a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018060200180602001806020018960ff1660ff1681526020018860ff1660ff1681526020018760ff1660ff168152602001868152602001856000191660001916815260200184810384528c818151815260200191508051906020019080838360005b838110156103a857808201518184015260208101905061038d565b50505050905090810190601f1680156103d55780820380516001836020036101000a031916815260200191505b5084810383528b818151815260200191508051906020019080838360005b8381101561040e5780820151818401526020810190506103f3565b50505050905090810190601f16801561043b5780820380516001836020036101000a031916815260200191505b5084810382528a818151815260200191508051906020019080838360005b83811015610474578082015181840152602081019050610459565b50505050905090810190601f1680156104a15780820380516001836020036101000a031916815260200191505b509c5050505050505050505050505060405180910390f35b600015156000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600b0160009054906101000a900460ff16151514151561051a57600080fd5b6101a0604051908101604052803373ffffffffffffffffffffffffffffffffffffffff1681526020018a81526020018981526020018881526020018760ff1681526020018660ff16815260200185600019168152602001600160ff168152602001848152602001838152602001828152602001428152602001600115158152506000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550602082015181600101908051906020019061063b929190610bd7565b506040820151816002019080519060200190610658929190610bd7565b506060820151816003019080519060200190610675929190610bd7565b5060808201518160040160006101000a81548160ff021916908360ff16021790555060a08201518160040160016101000a81548160ff021916908360ff16021790555060c0820151816005019060001916905560e08201518160060160006101000a81548160ff021916908360ff160217905550610100820151816007019080519060200190610706929190610bd7565b50610120820151816008019080519060200190610724929190610bd7565b50610140820151816009019080519060200190610742929190610bd7565b5061016082015181600a015561018082015181600b0160006101000a81548160ff0219169083151502179055509050507f6142ff54c11c1eb59d0a251a917b38ccd85c84dc4c83c021e8662a73d557b00560405160405180910390a1505050505050505050565b6000600115156000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600b0160009054906101000a900460ff16151514151561080c57600080fd5b6000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020905083816003019080519060200190610864929190610c57565b50828160040160016101000a81548160ff021916908360ff160217905550818160050181600019169055507f3c22cfbecb928778078fb52ac2ece267e23d58f14308b86f1645e39aaad0197860405160405180910390a150505050565b60006060806060600080600080600080600115156000808d73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600b0160009054906101000a900460ff16151514151561093257600080fd5b6000808c73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002090508060000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff168160010182600201836003018460040160009054906101000a900460ff168560040160019054906101000a900460ff168660060160009054906101000a900460ff1687600a01548860050154878054600181600116156101000203166002900480601f016020809104026020016040519081016040528092919081815260200182805460018160011615610100020316600290048015610a785780601f10610a4d57610100808354040283529160200191610a78565b820191906000526020600020905b815481529060010190602001808311610a5b57829003601f168201915b50505050509750868054600181600116156101000203166002900480601f016020809104026020016040519081016040528092919081815260200182805460018160011615610100020316600290048015610b145780601f10610ae957610100808354040283529160200191610b14565b820191906000526020600020905b815481529060010190602001808311610af757829003601f168201915b50505050509650858054600181600116156101000203166002900480601f016020809104026020016040519081016040528092919081815260200182805460018160011615610100020316600290048015610bb05780601f10610b8557610100808354040283529160200191610bb0565b820191906000526020600020905b815481529060010190602001808311610b9357829003601f168201915b50505050509550995099509950995099509950995099509950509193959799909294969850565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f10610c1857805160ff1916838001178555610c46565b82800160010185558215610c46579182015b82811115610c45578251825591602001919060010190610c2a565b5b509050610c539190610cd7565b5090565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f10610c9857805160ff1916838001178555610cc6565b82800160010185558215610cc6579182015b82811115610cc5578251825591602001919060010190610caa565b5b509050610cd39190610cd7565b5090565b610cf991905b80821115610cf5576000816000905550600101610cdd565b5090565b905600a165627a7a7230582003480c885cf69db002b75e4a0f99edbcfc997b3a6dec5d9826c3883503a7e93a0029";
var DPOSBallotContractAddress common.Address;

// provide SDK api to handle this.
type DAppManager interface {
	registerDAppInfo(dappName string, orgName string, orgDescription string, nationalityCode uint8, ledgerReplicated uint8, icon[] byte, ext0 string, ext1 string, ext2 string)
	updateDAppInfo(orgDescription string, ledgerReplicated uint8, icon[] byte)
	getDAppInfo()
}

// provide SDK api to handle this.
type DPoSBallot interface {
	register(nodeid string, ip string, port uint);
	vote(nodeid string);
	delegatorList();
	delegatorInfo(nodeid string);
}

/**
	nonce := uint64(0);
	gasLimit := config.GenesisGasLimit;
	gasPrice := big.NewInt(300000);
	data := []byte(DAPPContractBinCode);
	types.NewTransaction(nonce, common.Address{}, big.NewInt(0), gasLimit, gasPrice, data)
 */
