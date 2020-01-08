function upperCaseFirstLetter(string) {
    if(typeof string !== 'string') return string;
    string = string.replace(/^./, match => match.toUpperCase());
    return string;
}

function getParameterError({
    name = '',
    para,
    correct,
    wrong
}) {
    const parameter = para ? `parameter.${para}` : 'parameter';
    const errorType = upperCaseFirstLetter(wrong === null ? 'Null' : typeof wrong);
    return `${name}:fail parameter error: ${parameter} should be ${correct} instead of ${errorType}`;
}

function shouleBeObject(target) {
    if(target && typeof target === 'object') return { res: true };
    return {
        res: false,
        msg: getParameterError({
            correct: 'Object',
            wrong: target
        })
    };
}

export {
    shouleBeObject,
    getParameterError,
};
