// Script to generate 50 students data
const fs = require('fs');
const path = require('path');

// Spanish names for generation
const nombres = ['Juan', 'María', 'José', 'Ana', 'Carlos', 'Sofía', 'Miguel', 'Isabella', 'Pedro', 'Valentina', 'Andrés', 'Camila', 'Diego', 'Paula', 'Jorge', 'Mariana', 'Luis', 'Carolina', 'Francisco', 'Esperanza', 'Manuel', 'Lorena', 'Ricardo', 'Patricia', 'Fernando', 'Claudia', 'Alberto', 'Beatriz', 'Rafael', 'Gabriela', 'Sergio', 'Daniela', 'Pablo', 'Fernanda', 'Gustavo', 'Adriana', 'Roberto', 'Verónica', 'Eduardo', 'Alicia', 'Marcelo', 'Julia', 'Roberto', 'Natalia', 'Hugo', 'Milena', 'Santiago', 'Luciana', 'Antonio', 'Renata', 'Cristian', 'Florencia'];

const apellidos = ['Pérez', 'García', 'Rodríguez', 'López', 'González', 'Hernández', 'Martínez', 'Sánchez', 'Torres', 'Ramírez', 'Flores', 'Rivera', 'Gómez', 'Díaz', 'Reyes', 'Morales', 'Cruz', 'Ortiz', 'Gutierrez', 'Chávez', 'Ruiz', 'Vargas', 'Molina', 'Silva', 'Rojas', 'Medina', 'Castro', 'Huerta', 'Soto', 'Vega'];

const generos = ['Masculino', 'Femenino', 'Otro', 'Prefiero no decirlo'];
const codigosCarrera = ['10101', '10102', '10103'];
const eneatipos = ['e1', 'e2', 'e3', 'e4', 'e5', 'e6', 'e7', 'e8', 'e9'];

// Generate random Chilean RUT
function generarRUT() {
    const cuerpo = Math.floor(Math.random() * 99999999) + 1000000;
    const dv = Math.floor(Math.random() * 10);
    return `${cuerpo}-${dv}`;
}

// Generate random name from list
function randomFrom(arr) {
    return arr[Math.floor(Math.random() * arr.length)];
}

// Generate random float between min and max
function randomFloat(min, max) {
    return Math.round((Math.random() * (max - min) + min) * 100) / 100;
}

// Generate random social connections matrix
function generateSocialConnections(studentIds) {
    const connections = {};
    for (let i = 0; i < studentIds.length; i++) {
        connections[studentIds[i]] = {};
        for (let j = 0; j < studentIds.length; j++) {
            if (i !== j) {
                // Generate somewhat realistic connections
                // Same career or gender tends to have higher affinity
                connections[studentIds[i]][studentIds[j]] = Math.round((Math.random() * 2 - 1) * 100) / 100;
            }
        }
    }
    return connections;
}

// Generate OCEAN scores with some correlation to enneatype
function generateOCEAN(eneatipo) {
    // Base scores with some randomness
    const baseScores = {
        o: randomFloat(0.3, 0.9),
        c: randomFloat(0.3, 0.9),
        e: randomFloat(0.3, 0.9),
        a: randomFloat(0.3, 0.9),
        n: randomFloat(0.2, 0.8)
    };

    // Adjust based on enneatype for more realistic data
    const typeAdjustments = {
        e1: { c: 0.15, a: -0.1 },  // Type 1: more conscientious, less agreeable
        e2: { a: 0.2, n: 0.1 },     // Type 2: more agreeable, slightly more neurotic
        e3: { e: 0.15, n: -0.1 },   // Type 3: more extraverted, less neurotic
        e4: { o: 0.2, n: 0.1 },     // Type 4: more open, more neurotic
        e5: { o: 0.2, e: -0.15 },   // Type 5: more open, less extraverted
        e6: { a: 0.1, n: 0.1 },    // Type 6: more agreeable, more neurotic
        e7: { e: 0.15, c: -0.15 }, // Type 7: more extraverted, less conscientious
        e8: { a: -0.15, e: 0.1 },   // Type 8: less agreeable, more extraverted
        e9: { n: -0.15, e: -0.1 }   // Type 9: less neurotic, less extraverted
    };

    const adj = typeAdjustments[eneatipo] || {};
    for (const trait in adj) {
        baseScores[trait] = Math.max(0.1, Math.min(0.9, baseScores[trait] + adj[trait]));
    }

    return {
        openness: baseScores.o,
        conscientiousness: baseScores.c,
        extraversion: baseScores.e,
        agreeableness: baseScores.a,
        neuroticism: baseScores.n
    };
}

// Generate 50 students
const students = [];
const studentIds = [];

for (let i = 1; i <= 50; i++) {
    const id = `student_${String(i).padStart(3, '0')}`;
    studentIds.push(id);

    const nombresList = [randomFrom(nombres), randomFrom(nombres)];
    const paterno = randomFrom(apellidos);
    const materno = randomFrom(apellidos);
    const nombresFull = nombresList.join(' ');
    const rut = generarRUT();
    // Generate email without accents - use base name only
    const baseName = nombresList[0].toLowerCase().normalize("NFD").replace(/[\u0300-\u036f]/g, "");
    const basePaterno = paterno.toLowerCase().normalize("NFD").replace(/[\u0300-\u036f]/g, "");
    const email = `${baseName}.${basePaterno}${i}@alumnos.cl`;
    const career_code = randomFrom(codigosCarrera);
    const gender = randomFrom(generos);
    const eneatipo = randomFrom(eneatipos);
    const ocean = generateOCEAN(eneatipo);

    students.push({
        id,
        rut,
        nombres: nombresFull,
        paterno,
        materno,
        email,
        career_code,
        gender,
        eneatipo,
        ocean,
        social_connections: {} // Will be filled after all students are created
    });
}

// Generate social connections
const connections = generateSocialConnections(studentIds);
students.forEach(s => {
    s.social_connections = connections[s.id];
});

// Create final data structure
const data = {
    section_id: 3,
    students
};

// Write to file
const outputPath = path.join(__dirname, 'students_50.json');
fs.writeFileSync(outputPath, JSON.stringify(data, null, 2), 'utf8');
console.log(`Generated 50 students data to ${outputPath}`);
console.log(`Total students: ${students.length}`);
console.log(`Section ID: ${data.section_id}`);

// Also create a summary
console.log('\nSample student (first):');
console.log(JSON.stringify(students[0], null, 2));
